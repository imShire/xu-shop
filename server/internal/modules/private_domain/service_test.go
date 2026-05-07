package private_domain

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/errs"
	"github.com/xushop/xu-shop/internal/pkg/qywx"
	"github.com/xushop/xu-shop/internal/pkg/snowflake"
)

func init() {
	snowflake.Init(1)
}

// ---- mock ChannelCodeRepo ----

type mockChannelCodeRepo struct {
	codes     map[int64]*ChannelCode
	nameIndex map[string]int64 // name -> id
	idSeq     int64
}

func newMockChannelCodeRepo() *mockChannelCodeRepo {
	return &mockChannelCodeRepo{
		codes:     make(map[int64]*ChannelCode),
		nameIndex: make(map[string]int64),
	}
}

func (m *mockChannelCodeRepo) Create(_ context.Context, c *ChannelCode) error {
	if _, exists := m.nameIndex[c.Name]; exists {
		return errors.New("duplicate key value violates unique constraint")
	}
	m.codes[c.ID] = c
	m.nameIndex[c.Name] = c.ID
	return nil
}

func (m *mockChannelCodeRepo) FindByID(_ context.Context, id int64) (*ChannelCode, error) {
	if c, ok := m.codes[id]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, errors.New("record not found")
}

func (m *mockChannelCodeRepo) Update(_ context.Context, c *ChannelCode) error {
	m.codes[c.ID] = c
	return nil
}

func (m *mockChannelCodeRepo) Delete(_ context.Context, id int64) error {
	if _, ok := m.codes[id]; !ok {
		return errs.ErrNotFound
	}
	delete(m.codes, id)
	return nil
}

func (m *mockChannelCodeRepo) List(_ context.Context, _, _ int) ([]ChannelCode, int64, error) {
	return nil, 0, nil
}

// ---- mock TagRepo ----

type mockTagRepo struct {
	tags      map[int64]*CustomerTag
	nameIndex map[string]int64
}

func newMockTagRepo() *mockTagRepo {
	return &mockTagRepo{
		tags:      make(map[int64]*CustomerTag),
		nameIndex: make(map[string]int64),
	}
}

func (m *mockTagRepo) Create(_ context.Context, t *CustomerTag) error {
	if _, exists := m.nameIndex[t.Name]; exists {
		return errors.New("duplicate key value violates unique constraint \"customer_tag_name_key\"")
	}
	m.tags[t.ID] = t
	m.nameIndex[t.Name] = t.ID
	return nil
}

func (m *mockTagRepo) FindByID(_ context.Context, id int64) (*CustomerTag, error) {
	if t, ok := m.tags[id]; ok {
		cp := *t
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockTagRepo) Update(_ context.Context, t *CustomerTag) error {
	if existingID, exists := m.nameIndex[t.Name]; exists && existingID != t.ID {
		return errors.New("duplicate key value violates unique constraint \"customer_tag_name_key\"")
	}
	if current, exists := m.tags[t.ID]; exists {
		delete(m.nameIndex, current.Name)
	}
	cp := *t
	m.tags[t.ID] = &cp
	m.nameIndex[t.Name] = t.ID
	return nil
}

func (m *mockTagRepo) Delete(_ context.Context, id int64) error {
	delete(m.tags, id)
	return nil
}

func (m *mockTagRepo) List(_ context.Context) ([]CustomerTag, error) {
	var list []CustomerTag
	for _, t := range m.tags {
		list = append(list, *t)
	}
	return list, nil
}

// ---- mock UserTagRepo ----

type mockUserTagRepo struct {
	// key: "userID:tagID"
	entries  map[string]bool
	addCount int
}

func newMockUserTagRepo() *mockUserTagRepo {
	return &mockUserTagRepo{entries: make(map[string]bool)}
}

func (m *mockUserTagRepo) Add(_ context.Context, ut *UserTag) error {
	key := userTagKey(ut.UserID, ut.TagID)
	// Simulate ON CONFLICT DO NOTHING (idempotent)
	if !m.entries[key] {
		m.entries[key] = true
		m.addCount++
	}
	return nil
}

func (m *mockUserTagRepo) Remove(_ context.Context, userID, tagID int64) error {
	delete(m.entries, userTagKey(userID, tagID))
	return nil
}

func (m *mockUserTagRepo) ListByUser(_ context.Context, _ int64) ([]CustomerTag, error) {
	return nil, nil
}

func userTagKey(userID, tagID int64) string {
	return string(rune(userID)) + ":" + string(rune(tagID))
}

// ---- mock ShareRepo ----

type mockShareRepo struct{}

func (m *mockShareRepo) UpsertShortCode(_ context.Context, sc *ShareShortCode) (*ShareShortCode, error) {
	return sc, nil
}
func (m *mockShareRepo) FindShortCode(_ context.Context, _ int64) (*ShareShortCode, error) {
	return nil, errors.New("record not found")
}
func (m *mockShareRepo) CreateAttribution(_ context.Context, _ *ShareAttribution) error { return nil }
func (m *mockShareRepo) ExistsAttribution(_ context.Context, _, _, _ int64, _ time.Time) (bool, error) {
	return false, nil
}

// ---- helpers ----

func buildPrivateDomainSvc(
	channelRepo ChannelCodeRepo,
	tagRepo TagRepo,
	userTagRepo UserTagRepo,
) *Service {
	return NewService(channelRepo, tagRepo, userTagRepo, &mockShareRepo{}, qywx.NewMockClient(), nil)
}

// ---- Tests ----

// TestCreateChannelCode_DuplicateName 使用已存在名称创建渠道码时返回错误。
func TestCreateChannelCode_DuplicateName(t *testing.T) {
	channelRepo := newMockChannelCodeRepo()
	svc := buildPrivateDomainSvc(channelRepo, newMockTagRepo(), newMockUserTagRepo())

	req := CreateChannelCodeReq{
		Name:  "春节活动渠道",
		State: "state_001",
	}

	// 第一次创建成功
	_, err := svc.CreateChannelCode(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("first CreateChannelCode failed: %v", err)
	}

	// 第二次使用相同名称应返回错误
	_, err = svc.CreateChannelCode(context.Background(), 1, req)
	if err == nil {
		t.Fatal("expected error on duplicate channel code name, got nil")
	}
}

// TestDeleteChannelCode_NotFound 删除不存在的渠道码返回错误。
func TestDeleteChannelCode_NotFound(t *testing.T) {
	channelRepo := newMockChannelCodeRepo()
	svc := buildPrivateDomainSvc(channelRepo, newMockTagRepo(), newMockUserTagRepo())

	err := svc.DeleteChannelCode(context.Background(), 9999)
	if err == nil {
		t.Fatal("expected error when deleting non-existent channel code, got nil")
	}

	// 应返回 ErrNotFound 类型的错误
	ae, ok := err.(*errs.AppError)
	if !ok {
		t.Fatalf("expected AppError (ErrNotFound), got %T: %v", err, err)
	}
	if ae.Code != errs.ErrNotFound.Code {
		t.Errorf("expected ErrNotFound code=%d, got code=%d: %s", errs.ErrNotFound.Code, ae.Code, ae.Message)
	}
}

// TestCreateTag_Duplicate 使用重复名称创建标签时返回错误。
func TestCreateTag_Duplicate(t *testing.T) {
	tagRepo := newMockTagRepo()
	svc := buildPrivateDomainSvc(newMockChannelCodeRepo(), tagRepo, newMockUserTagRepo())

	req := CreateTagReq{
		Name:   "VIP客户",
		Source: "manual",
	}

	// 第一次创建成功
	if err := svc.CreateTag(context.Background(), req); err != nil {
		t.Fatalf("first CreateTag failed: %v", err)
	}

	// 第二次使用相同名称应返回错误
	err := svc.CreateTag(context.Background(), req)
	if err == nil {
		t.Fatal("expected error on duplicate tag name, got nil")
	}
	ae, ok := err.(*errs.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != errs.ErrConflict.Code {
		t.Fatalf("expected conflict code %d, got %d", errs.ErrConflict.Code, ae.Code)
	}
}

func TestCreateTag_StoresQYWXTagIDSeparately(t *testing.T) {
	tagRepo := newMockTagRepo()
	svc := buildPrivateDomainSvc(newMockChannelCodeRepo(), tagRepo, newMockUserTagRepo())

	if err := svc.CreateTag(context.Background(), CreateTagReq{
		Name:   "测试标签",
		Source: "manual",
	}); err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}

	var created *CustomerTag
	for _, tag := range tagRepo.tags {
		created = tag
		break
	}
	if created == nil {
		t.Fatal("expected created tag")
	}
	if created.Source != "manual" {
		t.Fatalf("expected source manual, got %s", created.Source)
	}
	if created.QYWXTagID == nil || *created.QYWXTagID == "" {
		t.Fatal("expected qywx_tag_id to be stored separately")
	}
}

func TestUpdateTag_Duplicate(t *testing.T) {
	tagRepo := newMockTagRepo()
	svc := buildPrivateDomainSvc(newMockChannelCodeRepo(), tagRepo, newMockUserTagRepo())

	first := &CustomerTag{ID: 1, Name: "VIP客户", Source: "manual"}
	second := &CustomerTag{ID: 2, Name: "新客", Source: "manual"}
	tagRepo.tags[first.ID] = first
	tagRepo.tags[second.ID] = second
	tagRepo.nameIndex[first.Name] = first.ID
	tagRepo.nameIndex[second.Name] = second.ID

	err := svc.UpdateTag(context.Background(), second.ID, UpdateTagReq{Name: first.Name})
	if err == nil {
		t.Fatal("expected duplicate error on update, got nil")
	}
	ae, ok := err.(*errs.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != errs.ErrConflict.Code {
		t.Fatalf("expected conflict code %d, got %d", errs.ErrConflict.Code, ae.Code)
	}
}

func TestUpdateTag_NotFound(t *testing.T) {
	svc := buildPrivateDomainSvc(newMockChannelCodeRepo(), newMockTagRepo(), newMockUserTagRepo())

	err := svc.UpdateTag(context.Background(), 999, UpdateTagReq{Name: "不存在"})
	if err == nil {
		t.Fatal("expected not found error, got nil")
	}
	ae, ok := err.(*errs.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != errs.ErrNotFound.Code {
		t.Fatalf("expected not found code %d, got %d", errs.ErrNotFound.Code, ae.Code)
	}
}

// TestAddUserTag_AlreadyExists 已分配的标签再次添加应幂等处理（不报错）。
func TestAddUserTag_AlreadyExists(t *testing.T) {
	userTagRepo := newMockUserTagRepo()
	svc := buildPrivateDomainSvc(newMockChannelCodeRepo(), newMockTagRepo(), userTagRepo)

	userID := int64(100)
	tagID := int64(1)

	// 第一次打标签
	if err := svc.AddUserTag(context.Background(), userID, tagID, "manual"); err != nil {
		t.Fatalf("first AddUserTag failed: %v", err)
	}

	// 第二次重复打相同标签，应幂等（不报错）
	if err := svc.AddUserTag(context.Background(), userID, tagID, "manual"); err != nil {
		t.Fatalf("second AddUserTag (idempotent) should not fail: %v", err)
	}

	// 实际写入次数只计一次（ON CONFLICT DO NOTHING）
	if userTagRepo.addCount != 1 {
		t.Errorf("expected 1 unique entry, got %d", userTagRepo.addCount)
	}
}
