package account_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/xushop/xu-shop/internal/config"
	"github.com/xushop/xu-shop/internal/modules/account"
	pkgjwt "github.com/xushop/xu-shop/internal/pkg/jwt"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
	"github.com/xushop/xu-shop/internal/pkg/wxlogin"
)

// ---- mock repos ----

type mockUserRepo struct {
	users map[string]*account.User // key=openid_mp
	byID  map[int64]*account.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[string]*account.User),
		byID:  make(map[int64]*account.User),
	}
}

func (m *mockUserRepo) FindByOpenidMP(_ context.Context, openid string) (*account.User, error) {
	if u, ok := m.users[openid]; ok {
		return u, nil
	}
	return nil, errors.New("record not found")
}
func (m *mockUserRepo) FindByOpenidH5(_ context.Context, _ string) (*account.User, error) {
	return nil, errors.New("record not found")
}
func (m *mockUserRepo) FindByID(_ context.Context, id int64) (*account.User, error) {
	if u, ok := m.byID[id]; ok {
		return u, nil
	}
	return nil, errors.New("record not found")
}
func (m *mockUserRepo) UpsertByOpenidMP(_ context.Context, u *account.User) error {
	if _, exists := m.users[*u.OpenidMP]; !exists {
		m.users[*u.OpenidMP] = u
		m.byID[u.ID] = u
	}
	return nil
}
func (m *mockUserRepo) UpsertByOpenidH5(_ context.Context, u *account.User) error {
	m.byID[u.ID] = u
	return nil
}
func (m *mockUserRepo) Update(_ context.Context, id int64, updates map[string]any) error {
	if u, ok := m.byID[id]; ok {
		if s, ok := updates["status"].(string); ok {
			u.Status = s
		}
	}
	return nil
}
func (m *mockUserRepo) CountByPhone(_ context.Context, _ string) (int64, error) { return 0, nil }
func (m *mockUserRepo) CountActiveByPhoneExclude(_ context.Context, _ string, _ int64) (int64, error) {
	return 0, nil
}
func (m *mockUserRepo) GetUserByPhone(_ context.Context, _ string) (*account.User, error) {
	return nil, errors.New("record not found")
}
func (m *mockUserRepo) SetPasswordHash(_ context.Context, _ int64, _ string) error { return nil }
func (m *mockUserRepo) CreateUser(_ context.Context, u *account.User) error {
	if _, exists := m.byID[u.ID]; !exists {
		m.byID[u.ID] = u
	}
	return nil
}
func (m *mockUserRepo) ListUsers(_ context.Context, _, _, _ string, _, _ int) ([]account.User, int64, error) {
	return nil, 0, nil
}
func (m *mockUserRepo) SessionKeyKey(_ int64) string { return "" }
func (m *mockUserRepo) GetBalance(_ context.Context, _ int64) (int64, error)  { return 0, nil }
func (m *mockUserRepo) RechargeBalance(_ context.Context, _ int64, _ int64, _ int64, _ string) error {
	return nil
}
func (m *mockUserRepo) DeductBalance(_ context.Context, _ int64, _ int64, _ string, _ int64, _ string) error {
	return nil
}
func (m *mockUserRepo) RefundBalance(_ context.Context, _ int64, _ int64, _ string, _ int64, _ string) error {
	return nil
}
func (m *mockUserRepo) ListBalanceLogs(_ context.Context, _ int64, _, _ int) ([]account.BalanceLog, int64, error) {
	return nil, 0, nil
}

// ---- mock admin repo ----

type mockAdminRepo struct {
	admins map[string]*account.Admin
	logs   []*account.LoginLog
}

func newMockAdminRepo() *mockAdminRepo {
	return &mockAdminRepo{admins: make(map[string]*account.Admin)}
}
func (m *mockAdminRepo) FindByUsername(_ context.Context, username string) (*account.Admin, error) {
	if a, ok := m.admins[username]; ok {
		return a, nil
	}
	return nil, errors.New("record not found")
}
func (m *mockAdminRepo) FindByID(_ context.Context, _ int64) (*account.Admin, error) {
	return nil, errors.New("not impl")
}
func (m *mockAdminRepo) Create(_ context.Context, a *account.Admin) error {
	if _, exists := m.admins[a.Username]; exists {
		return errors.New("duplicate")
	}
	m.admins[a.Username] = a
	return nil
}
func (m *mockAdminRepo) Update(_ context.Context, id int64, updates map[string]any) error {
	for _, a := range m.admins {
		if a.ID == id {
			if v, ok := updates["failed_attempts"].(int); ok {
				a.FailedAttempts = v
			}
			if v, ok := updates["locked_until"]; ok {
				if t, ok := v.(time.Time); ok {
					a.LockedUntil = &t
				} else {
					a.LockedUntil = nil
				}
			}
			if v, ok := updates["status"].(string); ok {
				a.Status = v
			}
		}
	}
	return nil
}
func (m *mockAdminRepo) List(_ context.Context, _, _ int) ([]account.Admin, int64, error) {
	list := make([]account.Admin, 0, len(m.admins))
	for _, admin := range m.admins {
		list = append(list, *admin)
	}
	return list, int64(len(list)), nil
}
func (m *mockAdminRepo) SaveLoginLog(_ context.Context, log *account.LoginLog) error {
	m.logs = append(m.logs, log)
	return nil
}

// ---- mock role repo ----

type mockRoleRepo struct{}

func (m *mockRoleRepo) ListRoles(_ context.Context) ([]account.Role, error) { return nil, nil }
func (m *mockRoleRepo) FindRoleByID(_ context.Context, _ int64) (*account.Role, error) {
	return nil, nil
}
func (m *mockRoleRepo) CreateRole(_ context.Context, _ *account.Role) error           { return nil }
func (m *mockRoleRepo) UpdateRole(_ context.Context, _ int64, _ map[string]any) error { return nil }
func (m *mockRoleRepo) DeleteRole(_ context.Context, _ int64) error                   { return nil }
func (m *mockRoleRepo) ListPermissions(_ context.Context) ([]account.Permission, error) {
	return nil, nil
}
func (m *mockRoleRepo) GetAdminPermCodes(_ context.Context, _ int64) ([]string, error) {
	return []string{"system.admin.view"}, nil
}
func (m *mockRoleRepo) GetAdminRoleCodes(_ context.Context, _ int64) ([]string, error) {
	return []string{"super_admin"}, nil
}
func (m *mockRoleRepo) SetAdminRoles(_ context.Context, _ int64, _ []int64) error { return nil }
func (m *mockRoleRepo) SetRolePerms(_ context.Context, _ int64, _ []string) error { return nil }

// ---- helpers ----

func newTestRdb(t *testing.T) *redis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb
}

func newTestCfg() *config.Config {
	cfg := &config.Config{}
	cfg.JWT.Secret = "test-secret-32-bytes-long-enough!"
	cfg.JWT.UserTTL = time.Hour
	cfg.JWT.AdminTTL = 8 * time.Hour
	return cfg
}

func jwtCfgFromCfg(cfg *config.Config) pkgjwt.Config {
	return pkgjwt.Config{
		Secret:      cfg.JWT.Secret,
		UserExpiry:  cfg.JWT.UserTTL,
		AdminExpiry: cfg.JWT.AdminTTL,
	}
}

func init() {
	snowflake.Init(1)
}

// ---- 测试用例 ----

func TestMpLogin_NewUser(t *testing.T) {
	rdb := newTestRdb(t)
	cfg := newTestCfg()
	userRepo := newMockUserRepo()
	svc := account.NewService(userRepo, newMockAdminRepo(), &mockRoleRepo{}, rdb, cfg,
		wxlogin.NewMockClient(), wxlogin.NewMockClient())

	result, err := svc.MpLogin(context.Background(), "fake-code")
	if err != nil {
		t.Fatalf("MpLogin error: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("expected non-empty access_token")
	}
	if result.UserID == 0 {
		t.Error("expected non-zero user_id")
	}

	// 验证 token 可解析
	jc := jwtCfgFromCfg(cfg)
	claims, err := pkgjwt.Parse(jc, result.AccessToken)
	if err != nil {
		t.Fatalf("parse access_token error: %v", err)
	}
	if claims.Sub != result.UserID {
		t.Errorf("token sub mismatch: want %d, got %d", result.UserID, claims.Sub)
	}
}

func TestMpLogin_ExistingUser(t *testing.T) {
	rdb := newTestRdb(t)
	cfg := newTestCfg()
	userRepo := newMockUserRepo()

	openid := "mock_openid_mp_001"
	existingID := int64(9999999)
	userRepo.users[openid] = &account.User{ID: existingID, OpenidMP: &openid, Status: "active"}
	userRepo.byID[existingID] = userRepo.users[openid]

	svc := account.NewService(userRepo, newMockAdminRepo(), &mockRoleRepo{}, rdb, cfg,
		wxlogin.NewMockClient(), wxlogin.NewMockClient())

	result, err := svc.MpLogin(context.Background(), "any-code")
	if err != nil {
		t.Fatalf("MpLogin error: %v", err)
	}
	if result.UserID != existingID {
		t.Errorf("expected userID=%d, got %d", existingID, result.UserID)
	}
}

func TestAdminLogin_Success(t *testing.T) {
	rdb := newTestRdb(t)
	cfg := newTestCfg()
	adminRepo := newMockAdminRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongP@ssw0rd!"), 12)
	adminRepo.admins["admin1"] = &account.Admin{
		ID:           1001,
		Username:     "admin1",
		PasswordHash: string(hash),
		Status:       "active",
	}

	svc := account.NewService(newMockUserRepo(), adminRepo, &mockRoleRepo{}, rdb, cfg,
		wxlogin.NewMockClient(), wxlogin.NewMockClient())

	// 先写验证码到 miniredis
	ctx := context.Background()
	_ = rdb.SetEx(ctx, "captcha:test-cap-1", "123456", 60*time.Second)

	resp, err := svc.AdminLogin(ctx, account.AdminLoginReq{
		Username:    "admin1",
		Password:    "StrongP@ssw0rd!",
		CaptchaID:   "test-cap-1",
		CaptchaCode: "123456",
	}, "127.0.0.1", "test-ua")
	if err != nil {
		t.Fatalf("AdminLogin error: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected access_token")
	}

	// 验证 token 携带正确 typ
	jc := jwtCfgFromCfg(cfg)
	claims, err := pkgjwt.Parse(jc, resp.AccessToken)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if claims.Typ != "admin" {
		t.Errorf("expected typ=admin, got %s", claims.Typ)
	}
}

func TestAdminCreateUser_DefaultsToPersonSource(t *testing.T) {
	rdb := newTestRdb(t)
	cfg := newTestCfg()
	userRepo := newMockUserRepo()
	svc := account.NewService(userRepo, newMockAdminRepo(), &mockRoleRepo{}, rdb, cfg,
		wxlogin.NewMockClient(), wxlogin.NewMockClient())

	resp, err := svc.AdminCreateUser(context.Background(), account.AdminCreateUserReq{
		Phone:    "13800138000",
		Password: "StrongP@ssw0rd!",
	})
	if err != nil {
		t.Fatalf("AdminCreateUser error: %v", err)
	}
	if resp == nil || resp.ID == 0 {
		t.Fatal("expected created user response")
	}

	created, ok := userRepo.byID[int64(resp.ID)]
	if !ok {
		t.Fatal("expected created user in repo")
	}
	if created.Source != "person" {
		t.Fatalf("expected source=person, got %s", created.Source)
	}
}

func TestListAdmins_PopulatesRolesAndPerms(t *testing.T) {
	rdb := newTestRdb(t)
	cfg := newTestCfg()
	adminRepo := newMockAdminRepo()
	adminRepo.admins["root"] = &account.Admin{
		ID:       1001,
		Username: "root",
		Status:   "active",
	}
	svc := account.NewService(newMockUserRepo(), adminRepo, &mockRoleRepo{}, rdb, cfg,
		wxlogin.NewMockClient(), wxlogin.NewMockClient())

	list, total, err := svc.ListAdmins(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("ListAdmins error: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 admin, got total=%d len=%d", total, len(list))
	}
	if len(list[0].Roles) == 0 || list[0].Roles[0] != "super_admin" {
		t.Fatalf("expected roles to be populated, got %#v", list[0].Roles)
	}
	if len(list[0].Perms) == 0 || list[0].Perms[0] != "system.admin.view" {
		t.Fatalf("expected perms to be populated, got %#v", list[0].Perms)
	}
}

func TestAdminLogin_WrongPwd(t *testing.T) {
	rdb := newTestRdb(t)
	cfg := newTestCfg()
	adminRepo := newMockAdminRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("CorrectP@ssw0rd!"), 12)
	adminRepo.admins["admin2"] = &account.Admin{
		ID: 1002, Username: "admin2", PasswordHash: string(hash), Status: "active",
	}
	svc := account.NewService(newMockUserRepo(), adminRepo, &mockRoleRepo{}, rdb, cfg,
		wxlogin.NewMockClient(), wxlogin.NewMockClient())

	ctx := context.Background()
	_ = rdb.SetEx(ctx, "captcha:cap-wrong", "654321", 60*time.Second)

	_, err := svc.AdminLogin(ctx, account.AdminLoginReq{
		Username: "admin2", Password: "WrongP@ssw0rd!",
		CaptchaID: "cap-wrong", CaptchaCode: "654321",
	}, "", "")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}

	// 验证失败次数增加
	admin := adminRepo.admins["admin2"]
	if admin.FailedAttempts != 1 {
		t.Errorf("expected FailedAttempts=1, got %d", admin.FailedAttempts)
	}
}

func TestAdminLogin_Locked(t *testing.T) {
	rdb := newTestRdb(t)
	cfg := newTestCfg()
	adminRepo := newMockAdminRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("P@ssw0rd123!"), 12)
	locked := time.Now().Add(10 * time.Minute)
	adminRepo.admins["admin3"] = &account.Admin{
		ID: 1003, Username: "admin3", PasswordHash: string(hash),
		Status: "active", LockedUntil: &locked,
	}
	svc := account.NewService(newMockUserRepo(), adminRepo, &mockRoleRepo{}, rdb, cfg,
		wxlogin.NewMockClient(), wxlogin.NewMockClient())

	ctx := context.Background()
	_ = rdb.SetEx(ctx, "captcha:cap-locked", "111111", 60*time.Second)

	_, err := svc.AdminLogin(ctx, account.AdminLoginReq{
		Username: "admin3", Password: "P@ssw0rd123!",
		CaptchaID: "cap-locked", CaptchaCode: "111111",
	}, "", "")
	if err == nil {
		t.Fatal("expected ErrAccountLocked, got nil")
	}
}

func TestAdminCreatePwdStrength(t *testing.T) {
	rdb := newTestRdb(t)
	cfg := newTestCfg()
	svc := account.NewService(newMockUserRepo(), newMockAdminRepo(), &mockRoleRepo{}, rdb, cfg,
		wxlogin.NewMockClient(), wxlogin.NewMockClient())

	// 正常创建
	resp, err := svc.CreateAdmin(context.Background(), account.CreateAdminReq{
		Username: "newadmin",
		Password: "StrongP@ssw0rd!",
		RealName: "Test Admin",
	})
	if err != nil {
		t.Fatalf("CreateAdmin error: %v", err)
	}
	if resp.ID == 0 {
		t.Error("expected non-zero admin ID")
	}
}

// 确保 jwt.RegisteredClaims 在 Claims 中正确使用
var _ = jwt.RegisteredClaims{}
