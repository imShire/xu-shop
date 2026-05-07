package shipping

import (
	"context"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"

	"github.com/xushop/xu-shop/internal/pkg/kdniao"
)

// ---- mock ShipmentRepo ----

type mockShipmentRepo struct {
	shipments map[int64]*Shipment
	tracks    map[int64][]ShipmentTrack
	idSeq     int64
}

func newMockShipmentRepo() *mockShipmentRepo {
	return &mockShipmentRepo{
		shipments: make(map[int64]*Shipment),
		tracks:    make(map[int64][]ShipmentTrack),
	}
}

func (m *mockShipmentRepo) nextID() int64 {
	m.idSeq++
	return m.idSeq
}

func (m *mockShipmentRepo) DB() *gorm.DB { return nil }

func (m *mockShipmentRepo) Create(_ context.Context, s *Shipment) error {
	if s.ID == 0 {
		s.ID = m.nextID()
	}
	cp := *s
	m.shipments[s.ID] = &cp
	return nil
}

func (m *mockShipmentRepo) FindByID(_ context.Context, id int64) (*Shipment, error) {
	if s, ok := m.shipments[id]; ok {
		cp := *s
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockShipmentRepo) FindByOrderID(_ context.Context, orderID int64) (*Shipment, error) {
	for _, s := range m.shipments {
		if s.OrderID == orderID {
			cp := *s
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockShipmentRepo) FindByCarrierAndNo(_ context.Context, carrierCode, trackingNo string) (*Shipment, error) {
	for _, s := range m.shipments {
		if s.CarrierCode == carrierCode && s.TrackingNo == trackingNo {
			cp := *s
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockShipmentRepo) Update(_ context.Context, s *Shipment) error {
	cp := *s
	m.shipments[s.ID] = &cp
	return nil
}

func (m *mockShipmentRepo) List(_ context.Context, _ ShipmentFilter) ([]Shipment, int64, error) {
	return nil, 0, nil
}

func (m *mockShipmentRepo) ListStale(_ context.Context, _ time.Time, _ []string) ([]Shipment, error) {
	return nil, nil
}

func (m *mockShipmentRepo) CreateTrack(_ context.Context, t *ShipmentTrack) error {
	if t.ID == 0 {
		t.ID = m.nextID()
	}
	cp := *t
	m.tracks[t.ShipmentID] = append(m.tracks[t.ShipmentID], cp)
	return nil
}

func (m *mockShipmentRepo) ListTracks(_ context.Context, shipmentID int64) ([]ShipmentTrack, error) {
	return m.tracks[shipmentID], nil
}

func (m *mockShipmentRepo) ExistsTrack(_ context.Context, shipmentID int64, occurredAt time.Time, description string) (bool, error) {
	for _, t := range m.tracks[shipmentID] {
		if t.OccurredAt.Equal(occurredAt) && t.Description != nil && *t.Description == description {
			return true, nil
		}
	}
	return false, nil
}

// ---- mock OrderAccessor ----

type mockOrderAccessorShip struct {
	orders      map[int64]*OrderSnap
	transitions []string
}

func (m *mockOrderAccessorShip) FindByID(_ context.Context, id int64) (*OrderSnap, error) {
	if o, ok := m.orders[id]; ok {
		cp := *o
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockOrderAccessorShip) Transition(_ context.Context, _ int64, trigger, _ string, _ int64, _ string) error {
	m.transitions = append(m.transitions, trigger)
	return nil
}

// ---- mock AsynqEnqueuer ----

type mockShipEnqueuer struct {
	tasks []*asynq.Task
}

func (m *mockShipEnqueuer) EnqueueContext(_ context.Context, t *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	m.tasks = append(m.tasks, t)
	return &asynq.TaskInfo{}, nil
}

// ---- 测试辅助 ----

func buildShipSvc(shipRepo ShipmentRepo, kdniaoClient kdniao.Client, orderDB OrderAccessor) *Service {
	return &Service{
		senderRepo:  &mockSenderAddressRepo{},
		carrierRepo: &mockCarrierRepo{},
		shipRepo:    shipRepo,
		kdniao:      kdniaoClient,
		orderDB:     orderDB,
		rdb:         nil,
		enqueuer:    nil,
		notifyURL:   "http://localhost/notify/express",
	}
}

func seededShipment(repo *mockShipmentRepo, id, orderID int64, carrierCode, trackingNo, status string) {
	now := time.Now()
	repo.shipments[id] = &Shipment{
		ID:               id,
		OrderID:          orderID,
		CarrierCode:      carrierCode,
		TrackingNo:       trackingNo,
		SenderSnapshot:   RawJSON("{}"),
		ReceiverSnapshot: RawJSON("{}"),
		Status:           status,
		LastTrackAt:      &now,
	}
}

// ---- mock SenderAddressRepo & CarrierRepo ----

type mockSenderAddressRepo struct{}

func (m *mockSenderAddressRepo) DB() *gorm.DB { return nil }
func (m *mockSenderAddressRepo) List(_ context.Context) ([]SenderAddress, error) {
	return nil, nil
}
func (m *mockSenderAddressRepo) FindByID(_ context.Context, _ int64) (*SenderAddress, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockSenderAddressRepo) FindDefault(_ context.Context) (*SenderAddress, error) {
	return &SenderAddress{
		ID: 1, Name: "测试发件人", Phone: "13800000000",
		Province: "广东", City: "深圳", District: "南山", Detail: "科技园南路",
	}, nil
}
func (m *mockSenderAddressRepo) Create(_ context.Context, _ *SenderAddress) error     { return nil }
func (m *mockSenderAddressRepo) Update(_ context.Context, _ *SenderAddress) error     { return nil }
func (m *mockSenderAddressRepo) Delete(_ context.Context, _ int64) error              { return nil }
func (m *mockSenderAddressRepo) SetDefault(_ context.Context, _ int64) error          { return nil }

type mockCarrierRepo struct{}

func (m *mockCarrierRepo) List(_ context.Context) ([]Carrier, error) { return nil, nil }
func (m *mockCarrierRepo) FindByCode(_ context.Context, code string) (*Carrier, error) {
	return &Carrier{Code: code, Name: "测试快递", KdniaoCode: "SF"}, nil
}
func (m *mockCarrierRepo) Update(_ context.Context, _ *Carrier) error { return nil }

// ---- 正式测试 ----

// TestHandleExpressPush_DeliveredState 推送 delivered 状态应更新发货单。
func TestHandleExpressPush_DeliveredState(t *testing.T) {
	shipRepo := newMockShipmentRepo()
	seededShipment(shipRepo, 1, 2001, "SF", "SF123456789", ShipStatusInTransit)

	now := time.Now()
	kdniaoMock := &kdniao.MockClient{
		PushResult: &kdniao.PushResp{
			CarrierCode: "SF",
			TrackingNo:  "SF123456789",
			State:       "delivered",
			Traces: []kdniao.TraceItem{
				{OccurredAt: now, Status: "delivered", Description: "已签收，感谢使用顺丰"},
			},
		},
	}
	orderDB := &mockOrderAccessorShip{
		orders: map[int64]*OrderSnap{
			2001: {ID: 2001, OrderNo: "ORDER2001", Status: "shipped"},
		},
	}

	svc := buildShipSvc(shipRepo, kdniaoMock, orderDB)

	// 模拟推送 body（具体内容由 ParsePush mock 忽略）
	err := svc.HandleExpressPush(context.Background(), []byte("mock_push_body"))
	if err != nil {
		t.Fatalf("HandleExpressPush failed: %v", err)
	}

	// 验证发货单状态已更新为 delivered
	ship := shipRepo.shipments[1]
	if ship.Status != ShipStatusDelivered {
		t.Fatalf("expected delivered, got %s", ship.Status)
	}
	if ship.DeliveredAt == nil {
		t.Fatal("delivered_at should be set")
	}

	// 验证轨迹已写入
	tracks := shipRepo.tracks[1]
	if len(tracks) == 0 {
		t.Fatal("shipment track should be created")
	}
	if tracks[0].Status != "delivered" {
		t.Fatalf("expected track status delivered, got %s", tracks[0].Status)
	}
}

// TestHandleExpressPush_Idempotent 相同轨迹推送两次，只写入一条记录。
func TestHandleExpressPush_Idempotent(t *testing.T) {
	shipRepo := newMockShipmentRepo()
	seededShipment(shipRepo, 2, 2002, "YTO", "YTO987654321", ShipStatusPicked)

	fixedTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.Local)
	traceDesc := "已从上海仓库发出"

	kdniaoMock := &kdniao.MockClient{
		PushResult: &kdniao.PushResp{
			CarrierCode: "YTO",
			TrackingNo:  "YTO987654321",
			State:       "in_transit",
			Traces: []kdniao.TraceItem{
				{OccurredAt: fixedTime, Status: "in_transit", Description: traceDesc},
			},
		},
	}
	orderDB := &mockOrderAccessorShip{
		orders: map[int64]*OrderSnap{
			2002: {ID: 2002, OrderNo: "ORDER2002", Status: "shipped"},
		},
	}

	svc := buildShipSvc(shipRepo, kdniaoMock, orderDB)

	// 第一次推送
	if err := svc.HandleExpressPush(context.Background(), []byte("push1")); err != nil {
		t.Fatal(err)
	}
	// 第二次推送（相同内容）
	if err := svc.HandleExpressPush(context.Background(), []byte("push2")); err != nil {
		t.Fatal(err)
	}

	// 验证轨迹只写入一条（幂等去重）
	tracks := shipRepo.tracks[2]
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track record (idempotent), got %d", len(tracks))
	}
	if tracks[0].Description == nil || *tracks[0].Description != traceDesc {
		t.Fatalf("expected description %q, got %v", traceDesc, tracks[0].Description)
	}
}
