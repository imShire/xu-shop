package account

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"log"
	"regexp"
	"time"

	"github.com/dchest/captcha"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/config"
	"github.com/xushop/xu-shop/internal/pkg/errs"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
	"github.com/xushop/xu-shop/internal/pkg/wxlogin"
)

const (
	// maxAdminLoginAttempts 连续失败次数上限。
	maxAdminLoginAttempts = 5
	// adminLockDuration 锁定时长。
	adminLockDuration = 15 * time.Minute
	// captchaTTL 验证码有效期。
	captchaTTL = 60 * time.Second
	// sessionKeyTTL 微信 session_key 缓存时长（2 小时）。
	sessionKeyTTL = 2 * time.Hour
	// deactivatePeriod 注销冷静期。
	deactivatePeriod = 30 * 24 * time.Hour
	// smsCodeTTL 短信验证码有效期（5 分钟）。
	smsCodeTTL = 5 * time.Minute
	// smsRateTTL 短信发送频率限制（60 秒内最多 1 次）。
	smsRateTTL = 60 * time.Second
	// loginFailMax 登录失败次数上限（10 分钟窗口）。
	loginFailMax = 10
	// loginFailTTL 登录失败计数窗口。
	loginFailTTL = 10 * time.Minute
	// loginLockDuration 登录锁定时长（30 分钟）。
	loginLockDuration = 30 * time.Minute
)

// phoneRegexp 手机号正则（11位数字，1开头）。
var phoneRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)

// Service 账号服务。
type Service struct {
	userRepo  UserRepo
	adminRepo AdminRepo
	roleRepo  RoleRepo
	rdb       *redis.Client
	jwtCfg    pkgjwt.Config
	wxMP      wxlogin.WxLoginClient // 小程序
	wxOA      wxlogin.WxLoginClient // 公众号
}

// NewService 构造 Service。
func NewService(
	userRepo UserRepo,
	adminRepo AdminRepo,
	roleRepo RoleRepo,
	rdb *redis.Client,
	cfg *config.Config,
	wxMP, wxOA wxlogin.WxLoginClient,
) *Service {
	return &Service{
		userRepo:  userRepo,
		adminRepo: adminRepo,
		roleRepo:  roleRepo,
		rdb:       rdb,
		jwtCfg: pkgjwt.Config{
			Secret:      cfg.JWT.Secret,
			UserExpiry:  cfg.JWT.UserTTL,
			AdminExpiry: cfg.JWT.AdminTTL,
		},
		wxMP: wxMP,
		wxOA: wxOA,
	}
}

// ---- C 端 ----

// MpLoginResult 小程序登录结果。
type MpLoginResult struct {
	AccessToken  string
	RefreshToken string
	UserID       int64
	ExpiresIn    int64
}

// MpLogin 小程序登录（code2session → upsert → 签发 JWT）。
func (s *Service) MpLogin(ctx context.Context, code string) (*MpLoginResult, error) {
	sess, err := s.wxMP.Code2Session(ctx, code)
	if err != nil {
		return nil, errs.ErrSessionExpired
	}

	id := snowflake.NextID()
	openid := sess.OpenID
	var unionid *string
	if sess.UnionID != "" {
		uid := sess.UnionID
		unionid = &uid
	}

	u := &User{
		ID:       id,
		OpenidMP: &openid,
		Unionid:  unionid,
		Status:   "active",
	}
	if err := s.userRepo.UpsertByOpenidMP(ctx, u); err != nil {
		return nil, errs.ErrInternal
	}

	// 重新查询真实记录（upsert 时 id 可能为已有值）
	found, err := s.userRepo.FindByOpenidMP(ctx, openid)
	if err != nil {
		return nil, errs.ErrInternal
	}

	// 缓存 session_key
	skKey := fmt.Sprintf("wx:sk:%d", found.ID)
	_ = s.rdb.SetEx(ctx, skKey, sess.SessionKey, sessionKeyTTL)

	return s.signUserToken(found.ID)
}

// H5GetOAuthURL 获取公众号 OAuth2 授权 URL。
func (s *Service) H5GetOAuthURL(_ context.Context, redirectURI, state string) string {
	return s.wxOA.GetOAuthURL(redirectURI, state)
}

// H5CallbackResult H5 回调登录结果。
type H5CallbackResult = MpLoginResult

// H5Callback 公众号 OAuth2 回调，upsert user，返回 token。
func (s *Service) H5Callback(ctx context.Context, code string) (*MpLoginResult, error) {
	resp, err := s.wxOA.OAuthCode2Token(ctx, code)
	if err != nil {
		return nil, errs.ErrSessionExpired
	}

	id := snowflake.NextID()
	openid := resp.OpenID
	var unionid *string
	if resp.UnionID != "" {
		uid := resp.UnionID
		unionid = &uid
	}

	u := &User{
		ID:       id,
		OpenidH5: &openid,
		Unionid:  unionid,
		Status:   "active",
	}
	if err := s.userRepo.UpsertByOpenidH5(ctx, u); err != nil {
		return nil, errs.ErrInternal
	}

	found, err := s.userRepo.FindByOpenidH5(ctx, openid)
	if err != nil {
		return nil, errs.ErrInternal
	}

	return s.signUserToken(found.ID)
}

// BindPhone 解密微信数据包并绑定手机号。
func (s *Service) BindPhone(ctx context.Context, userID int64, encryptedData, iv string) error {
	skKey := fmt.Sprintf("wx:sk:%d", userID)
	sessionKey, err := s.rdb.Get(ctx, skKey).Result()
	if err != nil {
		return errs.ErrSessionExpired
	}

	data, err := s.wxMP.DecryptUserData(sessionKey, encryptedData, iv)
	if err != nil {
		return errs.ErrParam.WithMsg("解密失败")
	}

	phone, ok := data["purePhoneNumber"].(string)
	if !ok || phone == "" {
		return errs.ErrParam.WithMsg("无法获取手机号")
	}

	// 检查手机号是否已被其他活跃账号绑定（排除当前用户自身）
	cnt, err := s.userRepo.CountActiveByPhoneExclude(ctx, phone, userID)
	if err != nil {
		return errs.ErrInternal
	}
	if cnt > 0 {
		return errs.ErrPhoneBound
	}

	return s.userRepo.Update(ctx, userID, map[string]any{"phone": phone})
}

// RefreshToken 使用 refresh token 换新 access token（一次性）。
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*MpLoginResult, error) {
	claims, err := pkgjwt.Parse(s.jwtCfg, refreshToken)
	if err != nil || claims.Typ != "user_refresh" {
		return nil, errs.ErrUnauth
	}

	// 检查黑名单
	exists, err := s.rdb.Exists(ctx, fmt.Sprintf("jwt:bl:%s", claims.JTI)).Result()
	if err == nil && exists > 0 {
		return nil, errs.ErrUnauth
	}

	// refresh token 一次性：立即加入黑名单
	if claims.ExpiresAt != nil {
		exp := time.Until(claims.ExpiresAt.Time)
		if exp > 0 {
			_ = s.rdb.SetEx(ctx, fmt.Sprintf("jwt:bl:%s", claims.JTI), "1", exp)
		}
	}

	return s.signUserToken(claims.Sub)
}

// Logout 将 JTI 加入 Redis 黑名单。
func (s *Service) Logout(ctx context.Context, jti string, exp time.Duration) error {
	if exp <= 0 {
		return nil
	}
	return s.rdb.SetEx(ctx, fmt.Sprintf("jwt:bl:%s", jti), "1", exp).Err()
}

// GetUser 查询用户信息。
func (s *Service) GetUser(ctx context.Context, userID int64) (*UserResp, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal
	}
	resp := toUserResp(u)
	return &resp, nil
}

// UpdateUser 更新用户资料。
func (s *Service) UpdateUser(ctx context.Context, userID int64, req UpdateUserReq) error {
	updates := map[string]any{}
	if req.Nickname != nil {
		updates["nickname"] = *req.Nickname
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	if req.Birthday != nil {
		updates["birthday"] = *req.Birthday
	}
	if len(updates) == 0 {
		return nil
	}
	return s.userRepo.Update(ctx, userID, updates)
}

// RequestDeactivate 申请注销，status=deactivating，deactivate_at=now+30d。
func (s *Service) RequestDeactivate(ctx context.Context, userID int64, _ string) error {
	deactivateAt := time.Now().Add(deactivatePeriod)
	return s.userRepo.Update(ctx, userID, map[string]any{
		"status":        "deactivating",
		"deactivate_at": deactivateAt,
	})
}

// CancelDeactivate 撤销注销申请。
func (s *Service) CancelDeactivate(ctx context.Context, userID int64) error {
	return s.userRepo.Update(ctx, userID, map[string]any{
		"status":        "active",
		"deactivate_at": nil,
	})
}

// ---- 手机号注册/登录 ----

// SendSmsCode 发送短信验证码（开发阶段打日志，不接真实 SMS）。
func (s *Service) SendSmsCode(ctx context.Context, phone, purpose string) error {
	// 频率限制：60s 内同一手机号最多 1 次
	rateKey := fmt.Sprintf("sms:rate:%s", phone)
	set, err := s.rdb.SetNX(ctx, rateKey, "1", smsRateTTL).Result()
	if err != nil {
		return errs.ErrInternal
	}
	if !set {
		return errs.ErrRateLimit.WithMsg("发送过于频繁，请 60 秒后重试")
	}

	// 生成 6 位随机数字验证码
	code, err := generateSmsCode()
	if err != nil {
		return errs.ErrInternal
	}

	// 存 Redis，TTL 5 分钟
	codeKey := fmt.Sprintf("sms:code:%s:%s", purpose, phone)
	if err := s.rdb.SetEx(ctx, codeKey, code, smsCodeTTL).Err(); err != nil {
		return errs.ErrInternal
	}

	// 开发阶段：打日志代替真实发送
	log.Printf("SMS to %s: %s", phone, code)
	return nil
}

// PhoneRegister 手机号注册。
func (s *Service) PhoneRegister(ctx context.Context, req PhoneRegisterReq) (*ClientLoginResp, error) {
	// 校验验证码
	if err := s.verifySmsCode(ctx, req.Phone, "register", req.Code); err != nil {
		return nil, err
	}

	// 密码强度校验
	if err := validatePassword(req.Password); err != nil {
		return nil, errs.ErrParam.WithMsg(err.Error())
	}

	// 手机号未注册（active 状态查重）
	cnt, err := s.userRepo.CountByPhone(ctx, req.Phone)
	if err != nil {
		return nil, errs.ErrInternal
	}
	if cnt > 0 {
		return nil, errs.ErrConflict.WithMsg("该手机号已注册")
	}

	// bcrypt hash 密码
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, errs.ErrInternal
	}

	// 创建用户（nickname 默认 "用户" + 手机后4位，source = "phone"）
	phone := req.Phone
	hashStr := string(hash)
	suffix := phone
	if len(phone) >= 4 {
		suffix = phone[len(phone)-4:]
	}
	nickname := "用户" + suffix
	source := "phone"
	u := &User{
		ID:           snowflake.NextID(),
		Phone:        &phone,
		PasswordHash: &hashStr,
		Source:       source,
		Nickname:     &nickname,
		Status:       "active",
	}
	if err := s.userRepo.CreateUser(ctx, u); err != nil {
		return nil, errs.ErrInternal
	}

	// 创建成功后删除 Redis 验证码
	codeKey := fmt.Sprintf("sms:code:register:%s", req.Phone)
	_ = s.rdb.Del(ctx, codeKey)

	// 颁发 JWT
	result, err := s.signUserToken(u.ID)
	if err != nil {
		return nil, err
	}

	userResp := toUserResp(u)
	return &ClientLoginResp{
		Token: result.AccessToken,
		User:  userResp,
	}, nil
}

// PhoneLogin 手机号密码登录。
func (s *Service) PhoneLogin(ctx context.Context, req PhoneLoginReq) (*ClientLoginResp, error) {
	// 检查锁定状态
	lockKey := fmt.Sprintf("login:lock:%s", req.Phone)
	locked, err := s.rdb.Exists(ctx, lockKey).Result()
	if err != nil {
		return nil, errs.ErrInternal
	}
	if locked > 0 {
		return nil, errs.ErrAccountLocked.WithMsg("账号已锁定，请 30 分钟后重试")
	}

	// 查询用户
	u, err := s.userRepo.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.incLoginFail(ctx, req.Phone)
			return nil, errs.ErrParam.WithMsg("手机号或密码错误")
		}
		return nil, errs.ErrInternal
	}

	// 校验密码
	if u.PasswordHash == nil {
		s.incLoginFail(ctx, req.Phone)
		return nil, errs.ErrParam.WithMsg("该账号未设置密码，请使用其他方式登录")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(req.Password)); err != nil {
		s.incLoginFail(ctx, req.Phone)
		return nil, errs.ErrParam.WithMsg("手机号或密码错误")
	}

	// 登录成功，清除失败计数
	failKey := fmt.Sprintf("login:fail:%s", req.Phone)
	_ = s.rdb.Del(ctx, failKey)

	// 检查账号状态
	if u.Status != "active" {
		return nil, errs.ErrAccountDisabled
	}

	// 颁发 JWT
	result, err := s.signUserToken(u.ID)
	if err != nil {
		return nil, err
	}

	userResp := toUserResp(u)
	return &ClientLoginResp{
		Token: result.AccessToken,
		User:  userResp,
	}, nil
}

// ResetPassword 通过短信验证码重置密码。
func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordReq) error {
	// 校验验证码
	if err := s.verifySmsCode(ctx, req.Phone, "reset", req.Code); err != nil {
		return err
	}

	// 密码强度校验
	if err := validatePassword(req.NewPassword); err != nil {
		return errs.ErrParam.WithMsg(err.Error())
	}

	// 查询用户
	u, err := s.userRepo.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.ErrNotFound.WithMsg("手机号未注册")
		}
		return errs.ErrInternal
	}

	// 更新密码
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		return errs.ErrInternal
	}
	if err := s.userRepo.SetPasswordHash(ctx, u.ID, string(hash)); err != nil {
		return err
	}

	// 更新成功后删除 Redis 验证码
	codeKey := fmt.Sprintf("sms:code:reset:%s", req.Phone)
	_ = s.rdb.Del(ctx, codeKey)
	return nil
}

// verifySmsCode 校验短信验证码（校验后不删除，由调用方决定是否删除）。
func (s *Service) verifySmsCode(ctx context.Context, phone, purpose, code string) error {
	codeKey := fmt.Sprintf("sms:code:%s:%s", purpose, phone)
	stored, err := s.rdb.Get(ctx, codeKey).Result()
	if err != nil {
		return errs.ErrParam.WithMsg("验证码已过期或不存在")
	}
	if stored != code {
		return errs.ErrParam.WithMsg("验证码错误")
	}
	return nil
}

// incLoginFail 增加登录失败计数，超限则锁定。
func (s *Service) incLoginFail(ctx context.Context, phone string) {
	failKey := fmt.Sprintf("login:fail:%s", phone)
	pipe := s.rdb.Pipeline()
	incrCmd := pipe.Incr(ctx, failKey)
	pipe.Expire(ctx, failKey, loginFailTTL)
	_, _ = pipe.Exec(ctx)
	cnt := incrCmd.Val()
	if cnt >= loginFailMax {
		lockKey := fmt.Sprintf("login:lock:%s", phone)
		_ = s.rdb.SetEx(ctx, lockKey, "1", loginLockDuration)
		_ = s.rdb.Del(ctx, failKey)
	}
}

// generateSmsCode 生成 6 位随机数字验证码。
func generateSmsCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// 取无符号 32 位整数，对 1000000 取模保证 6 位
	n := (uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])) % 1000000
	return fmt.Sprintf("%06d", n), nil
}

// validateAdminPassword 校验管理员密码强度（≥12 位，同时包含字母、数字和特殊符号）。
func validateAdminPassword(pwd string) error {
	if len(pwd) < 12 {
		return errors.New("管理员密码至少 12 位")
	}
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(pwd)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(pwd)
	hasSpecial := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(pwd)
	if !hasLetter || !hasDigit || !hasSpecial {
		return errors.New("管理员密码需同时包含字母、数字和特殊符号")
	}
	return nil
}

// validatePassword 校验密码强度（≥8位，含字母、数字和特殊符号）。
func validatePassword(pwd string) error {
	if len(pwd) < 8 {
		return errors.New("密码至少8位")
	}
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(pwd)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(pwd)
	hasSpecial := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(pwd)
	if !hasLetter || !hasDigit || !hasSpecial {
		return errors.New("密码须包含字母、数字和特殊符号")
	}
	return nil
}

// ---- Admin 端 ----

// AdminGetCaptcha 生成图形验证码，存 Redis 60s，返回 captcha_id + base64 PNG。
func (s *Service) AdminGetCaptcha(ctx context.Context) (*CaptchaResp, error) {
	id := uuid.NewString()
	digits := captcha.RandomDigits(6)
	img := captcha.NewImage(id, digits, captcha.StdWidth, captcha.StdHeight)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, errs.ErrInternal
	}

	// 将数字数组转成字符串存 Redis
	var code string
	for _, d := range digits {
		code += fmt.Sprintf("%d", d)
	}
	if err := s.rdb.SetEx(ctx, "captcha:"+id, code, captchaTTL).Err(); err != nil {
		return nil, errs.ErrServiceDegraded
	}

	return &CaptchaResp{
		CaptchaID: id,
		ImageB64:  "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}

// AdminLogin 管理员登录：验证码 → bcrypt → 锁定逻辑 → login_log → 签发 JWT。
func (s *Service) AdminLogin(ctx context.Context, req AdminLoginReq, ip, ua string) (*TokenResp, error) {
	// 校验验证码（GetDel 消费一次）
	stored, err := s.rdb.GetDel(ctx, "captcha:"+req.CaptchaID).Result()
	if err != nil || stored != req.CaptchaCode {
		return nil, errs.ErrParam.WithMsg("验证码错误或已过期")
	}

	admin, err := s.adminRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrParam.WithMsg("用户名或密码错误")
		}
		return nil, errs.ErrInternal
	}

	// 检查锁定状态
	if admin.LockedUntil != nil && admin.LockedUntil.After(time.Now()) {
		s.saveLoginLog(ctx, "admin", &admin.ID, ip, ua, false, "账号锁定中")
		return nil, errs.ErrAccountLocked
	}

	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		attempts := admin.FailedAttempts + 1
		updates := map[string]any{"failed_attempts": attempts}
		if attempts >= maxAdminLoginAttempts {
			lockedUntil := time.Now().Add(adminLockDuration)
			updates["locked_until"] = lockedUntil
			updates["failed_attempts"] = 0
		}
		_ = s.adminRepo.Update(ctx, admin.ID, updates)
		s.saveLoginLog(ctx, "admin", &admin.ID, ip, ua, false, "密码错误")
		return nil, errs.ErrParam.WithMsg("用户名或密码错误")
	}

	// 重置失败计数
	_ = s.adminRepo.Update(ctx, admin.ID, map[string]any{
		"failed_attempts": 0,
		"locked_until":    nil,
		"last_login_at":   time.Now(),
		"last_login_ip":   ip,
	})

	// 查询权限点
	permCodes, _ := s.roleRepo.GetAdminPermCodes(ctx, admin.ID)
	roleCodes, _ := s.roleRepo.GetAdminRoleCodes(ctx, admin.ID)

	jti := uuid.NewString()
	claims := pkgjwt.Claims{
		Sub:   admin.ID,
		Typ:   "admin",
		JTI:   jti,
		Roles: roleCodes,
		Perms: permCodes,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtCfg.AdminExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := pkgjwt.Sign(s.jwtCfg, claims)
	if err != nil {
		return nil, errs.ErrInternal
	}

	s.saveLoginLog(ctx, "admin", &admin.ID, ip, ua, true, "")

	return &TokenResp{
		AccessToken: token,
		ExpiresIn:   int64(s.jwtCfg.AdminExpiry.Seconds()),
	}, nil
}

// AdminLogout 将 JTI 加入黑名单。
func (s *Service) AdminLogout(ctx context.Context, jti string, exp time.Duration) error {
	return s.Logout(ctx, jti, exp)
}

// CreateAdmin 创建管理员（密码强度校验 + bcrypt hash）。
func (s *Service) CreateAdmin(ctx context.Context, req CreateAdminReq) (*AdminResp, error) {
	if err := validateAdminPassword(req.Password); err != nil {
		return nil, errs.ErrParam.WithMsg(err.Error())
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, errs.ErrInternal
	}

	admin := &Admin{
		ID:           snowflake.NextID(),
		Username:     req.Username,
		PasswordHash: string(hash),
		Status:       "active",
	}
	if req.RealName != "" {
		admin.RealName = &req.RealName
	}
	admin.Phone = req.Phone

	if err := s.adminRepo.Create(ctx, admin); err != nil {
		return nil, errs.ErrConflict.WithMsg("用户名已存在")
	}

	if len(req.RoleIDs) > 0 {
		roleIDs := make([]int64, len(req.RoleIDs))
		for i, v := range req.RoleIDs {
			roleIDs[i] = v.Int64()
		}
		_ = s.roleRepo.SetAdminRoles(ctx, admin.ID, roleIDs)
	}

	resp := toAdminResp(admin)
	return &resp, nil
}

// UpdateAdmin 更新管理员信息。
func (s *Service) UpdateAdmin(ctx context.Context, id int64, req UpdateAdminReq) error {
	updates := map[string]any{}
	if req.RealName != nil {
		updates["real_name"] = *req.RealName
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if len(updates) > 0 {
		if err := s.adminRepo.Update(ctx, id, updates); err != nil {
			return errs.ErrInternal
		}
	}
	if req.RoleIDs != nil {
		roleIDs := make([]int64, len(req.RoleIDs))
		for i, v := range req.RoleIDs {
			roleIDs[i] = v.Int64()
		}
		_ = s.roleRepo.SetAdminRoles(ctx, id, roleIDs)
	}
	return nil
}

// DisableAdmin 禁用管理员。
func (s *Service) DisableAdmin(ctx context.Context, id int64) error {
	if err := s.adminRepo.Update(ctx, id, map[string]any{"status": "disabled"}); err != nil {
		return err
	}
	_ = s.rdb.Del(ctx, fmt.Sprintf("admin:status:%d", id))
	return nil
}

// EnableAdmin 启用管理员。
func (s *Service) EnableAdmin(ctx context.Context, id int64) error {
	if err := s.adminRepo.Update(ctx, id, map[string]any{"status": "active"}); err != nil {
		return err
	}
	_ = s.rdb.Del(ctx, fmt.Sprintf("admin:status:%d", id))
	return nil
}

// ResetAdminPwd 重置管理员密码（强度校验 + bcrypt）。
func (s *Service) ResetAdminPwd(ctx context.Context, id int64, newPwd string) error {
	if err := validateAdminPassword(newPwd); err != nil {
		return errs.ErrParam.WithMsg(err.Error())
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), 12)
	if err != nil {
		return errs.ErrInternal
	}
	return s.adminRepo.Update(ctx, id, map[string]any{"password_hash": string(hash)})
}

// GetAdminMe 查询当前管理员信息。
func (s *Service) GetAdminMe(ctx context.Context, adminID int64) (*AdminResp, error) {
	a, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, errs.ErrNotFound
	}
	resp := toAdminResp(a)
	resp.Roles, _ = s.roleRepo.GetAdminRoleCodes(ctx, adminID)
	resp.Perms, _ = s.roleRepo.GetAdminPermCodes(ctx, adminID)
	return &resp, nil
}

// ListAdmins 分页查询管理员列表。
func (s *Service) ListAdmins(ctx context.Context, page, pageSize int) ([]AdminResp, int64, error) {
	list, total, err := s.adminRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	resp := make([]AdminResp, len(list))
	for i, a := range list {
		resp[i] = toAdminResp(&a)
		resp[i].Roles, _ = s.roleRepo.GetAdminRoleCodes(ctx, a.ID)
		resp[i].Perms, _ = s.roleRepo.GetAdminPermCodes(ctx, a.ID)
	}
	return resp, total, nil
}

// ListRoles 查询角色列表。
func (s *Service) ListRoles(ctx context.Context) ([]RoleResp, error) {
	roles, err := s.roleRepo.ListRoles(ctx)
	if err != nil {
		return nil, errs.ErrInternal
	}
	resp := make([]RoleResp, len(roles))
	for i, r := range roles {
		resp[i] = toRoleResp(&r)
	}
	return resp, nil
}

// CreateRole 创建角色。
func (s *Service) CreateRole(ctx context.Context, req CreateRoleReq) (*RoleResp, error) {
	perms := make([]Permission, len(req.PermCodes))
	for i, code := range req.PermCodes {
		perms[i] = Permission{Code: code}
	}
	role := &Role{
		ID:          snowflake.NextID(),
		Code:        req.Code,
		Name:        req.Name,
		Permissions: perms,
	}
	if err := s.roleRepo.CreateRole(ctx, role); err != nil {
		return nil, errs.ErrConflict.WithMsg("角色 code 已存在")
	}
	resp := toRoleResp(role)
	return &resp, nil
}

// UpdateRole 更新角色。
func (s *Service) UpdateRole(ctx context.Context, id int64, req UpdateRoleReq) error {
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if len(updates) > 0 {
		if err := s.roleRepo.UpdateRole(ctx, id, updates); err != nil {
			return errs.ErrInternal
		}
	}
	if req.PermCodes != nil {
		if err := s.roleRepo.SetRolePerms(ctx, id, req.PermCodes); err != nil {
			return errs.ErrInternal
		}
	}
	return nil
}

// DeleteRole 删除角色（系统角色不可删）。
func (s *Service) DeleteRole(ctx context.Context, id int64) error {
	if err := s.roleRepo.DeleteRole(ctx, id); err != nil {
		return errs.ErrForbidden.WithMsg(err.Error())
	}
	return nil
}

// ListPermissions 查询所有权限点。
func (s *Service) ListPermissions(ctx context.Context) ([]PermissionResp, error) {
	perms, err := s.roleRepo.ListPermissions(ctx)
	if err != nil {
		return nil, errs.ErrInternal
	}
	resp := make([]PermissionResp, len(perms))
	for i, p := range perms {
		resp[i] = PermissionResp{Code: p.Code, Module: p.Module, Action: p.Action, Name: p.Name, Group: p.Module}
	}
	return resp, nil
}

// ---- 内部工具 ----

// signUserToken 签发 user access token + refresh token。
func (s *Service) signUserToken(userID int64) (*MpLoginResult, error) {
	jti := uuid.NewString()
	claims := pkgjwt.Claims{
		Sub: userID,
		Typ: "user",
		JTI: jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtCfg.UserExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken, err := pkgjwt.Sign(s.jwtCfg, claims)
	if err != nil {
		return nil, errs.ErrInternal
	}

	// refresh token 有效期为 access token 的 3 倍
	rjti := uuid.NewString()
	rClaims := pkgjwt.Claims{
		Sub: userID,
		Typ: "user_refresh",
		JTI: rjti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtCfg.UserExpiry * 3)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken, err := pkgjwt.Sign(s.jwtCfg, rClaims)
	if err != nil {
		return nil, errs.ErrInternal
	}

	return &MpLoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       userID,
		ExpiresIn:    int64(s.jwtCfg.UserExpiry.Seconds()),
	}, nil
}

func (s *Service) saveLoginLog(ctx context.Context, subjectType string, subjectID *int64, ip, ua string, success bool, failReason string) {
	log := &LoginLog{
		ID:          snowflake.NextID(),
		SubjectType: subjectType,
		SubjectID:   subjectID,
		Success:     success,
	}
	if ip != "" {
		log.IP = &ip
	}
	if ua != "" {
		log.UA = &ua
	}
	if !success && failReason != "" {
		log.FailReason = &failReason
	}
	_ = s.adminRepo.SaveLoginLog(ctx, log)
}

// AdminListUsers 管理员分页查询 C 端用户。
func (s *Service) AdminListUsers(ctx context.Context, q AdminListUsersQuery) ([]UserResp, int64, error) {
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}
	users, total, err := s.userRepo.ListUsers(ctx, q.Phone, q.Nickname, q.Status, page, pageSize)
	if err != nil {
		return nil, 0, errs.ErrInternal
	}
	resp := make([]UserResp, len(users))
	for i, u := range users {
		resp[i] = toUserResp(&u)
	}
	return resp, total, nil
}

// AdminCreateUser 管理员创建 C 端用户（手机号+密码，无需短信验证码）。
func (s *Service) AdminCreateUser(ctx context.Context, req AdminCreateUserReq) (*UserResp, error) {
	if !regexp.MustCompile(`^1[3-9]\d{9}$`).MatchString(req.Phone) {
		return nil, errs.ErrParam.WithMsg("手机号格式错误")
	}
	if err := validatePassword(req.Password); err != nil {
		return nil, errs.ErrParam.WithMsg(err.Error())
	}
	cnt, err := s.userRepo.CountByPhone(ctx, req.Phone)
	if err != nil {
		return nil, errs.ErrInternal
	}
	if cnt > 0 {
		return nil, errs.ErrConflict.WithMsg("该手机号已注册")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, errs.ErrInternal
	}
	phone := req.Phone
	hashStr := string(hash)
	source := "person"
	nickname := req.Nickname
	if nickname == nil {
		suffix := phone
		if len(phone) >= 4 {
			suffix = phone[len(phone)-4:]
		}
		n := "用户" + suffix
		nickname = &n
	}
	u := &User{
		ID:           snowflake.NextID(),
		Phone:        &phone,
		PasswordHash: &hashStr,
		Nickname:     nickname,
		Source:       source,
		Status:       "active",
	}
	if err := s.userRepo.CreateUser(ctx, u); err != nil {
		return nil, errs.ErrInternal
	}
	resp := toUserResp(u)
	return &resp, nil
}

// AdminGetUser 管理员查询单个 C 端用户详情（含余额）。
func (s *Service) AdminGetUser(ctx context.Context, userID int64) (*AdminUserDetailResp, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errs.ErrNotFound
	}
	resp := toUserResp(u)
	return &resp, nil
}

// AdminDisableUser 管理员禁用 C 端用户。
func (s *Service) AdminDisableUser(ctx context.Context, userID int64) error {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errs.ErrNotFound
	}
	if u.Status == "disabled" {
		return nil
	}
	return s.userRepo.Update(ctx, userID, map[string]any{"status": "disabled"})
}

// AdminEnableUser 管理员启用 C 端用户。
func (s *Service) AdminEnableUser(ctx context.Context, userID int64) error {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errs.ErrNotFound
	}
	if u.Status == "active" {
		return nil
	}
	return s.userRepo.Update(ctx, userID, map[string]any{"status": "active"})
}

// ---- 余额 ----

// GetBalance 获取用户余额。
func (s *Service) GetBalance(ctx context.Context, userID int64) (int64, error) {
	return s.userRepo.GetBalance(ctx, userID)
}

// RechargeBalance Admin 充值用户余额。
func (s *Service) RechargeBalance(ctx context.Context, userID, amountCents, operatorID int64, remark string) error {
	if amountCents <= 0 {
		return errs.ErrParam.WithMsg("充值金额必须大于 0")
	}
	return s.userRepo.RechargeBalance(ctx, userID, amountCents, operatorID, remark)
}

// ListBalanceLogs 余额流水列表。
func (s *Service) ListBalanceLogs(ctx context.Context, userID int64, page, size int) ([]BalanceLog, int64, error) {
	return s.userRepo.ListBalanceLogs(ctx, userID, page, size)
}
