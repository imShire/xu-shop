import type { CartItem } from '@/types/biz'

export interface ShowcaseHeroSlide {
  id: number
  image: string
  badge: string
  title: string
  subtitle: string
}

export interface ShowcaseQuickEntry {
  id: number
  label: string
  image: string
}

export interface ShowcaseCategoryItem {
  id: number
  label: string
  image: string
  productId: number
}

export interface ShowcaseDepartment {
  id: number
  name: string
  heroImage: string
  items: ShowcaseCategoryItem[]
}

export const mallSearchTotal = 238

export const showcaseHeroSlides: ShowcaseHeroSlide[] = [
  {
    id: 1,
    image: 'https://images.unsplash.com/photo-1517248135467-4c7edcad34c4?auto=format&fit=crop&w=1200&q=80',
    badge: '鲜果冻·免费尝鲜',
    title: '食物背后的美味人生',
    subtitle: '初夏人气零食相伴 ¥12.9 起',
  },
  {
    id: 2,
    image: 'https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=1200&q=80',
    badge: '生活馆上新',
    title: '把日用和餐厨一起备齐',
    subtitle: '从轻家居到零食补给，一次逛完',
  },
]

export const showcaseQuickEntries: ShowcaseQuickEntry[] = [
  {
    id: 1,
    label: '居家',
    image: 'https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?auto=format&fit=crop&w=240&q=80',
  },
  {
    id: 2,
    label: '餐厨',
    image: 'https://images.unsplash.com/photo-1556911220-bff31c812dba?auto=format&fit=crop&w=240&q=80',
  },
  {
    id: 3,
    label: '饮食',
    image: 'https://images.unsplash.com/photo-1547592166-23ac45744acd?auto=format&fit=crop&w=240&q=80',
  },
  {
    id: 4,
    label: '配件',
    image: 'https://images.unsplash.com/photo-1521572267360-ee0c2909d518?auto=format&fit=crop&w=240&q=80',
  },
  {
    id: 5,
    label: '服装',
    image: 'https://images.unsplash.com/photo-1445205170230-053b83016050?auto=format&fit=crop&w=240&q=80',
  },
  {
    id: 6,
    label: '婴童',
    image: 'https://images.unsplash.com/photo-1515488042361-ee00e0ddd4e4?auto=format&fit=crop&w=240&q=80',
  },
  {
    id: 7,
    label: '杂货',
    image: 'https://images.unsplash.com/photo-1517142089942-ba376ce32a2e?auto=format&fit=crop&w=240&q=80',
  },
  {
    id: 8,
    label: '洗护',
    image: 'https://images.unsplash.com/photo-1526947425960-945c6e72858f?auto=format&fit=crop&w=240&q=80',
  },
  {
    id: 9,
    label: '志趣',
    image: 'https://images.unsplash.com/photo-1516979187457-637abb4f9353?auto=format&fit=crop&w=240&q=80',
  },
]

export const showcaseDepartments: ShowcaseDepartment[] = [
  {
    id: 1,
    name: '居家',
    heroImage: 'https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?auto=format&fit=crop&w=960&q=80',
    items: [
      { id: 101, label: '布艺软装', image: 'https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?auto=format&fit=crop&w=360&q=80', productId: 101 },
      { id: 102, label: '被枕', image: 'https://images.unsplash.com/photo-1505693416388-c5ce068fe85?auto=format&fit=crop&w=360&q=80', productId: 102 },
      { id: 103, label: '床品件套', image: 'https://images.unsplash.com/photo-1505693416388-8f5ce068fe85?auto=format&fit=crop&w=360&q=80', productId: 103 },
      { id: 104, label: '灯具', image: 'https://images.unsplash.com/photo-1507473885765-e6ed057f782c?auto=format&fit=crop&w=360&q=80', productId: 101 },
      { id: 105, label: '地垫', image: 'https://images.unsplash.com/photo-1505693416388-36fce068fe85?auto=format&fit=crop&w=360&q=80', productId: 102 },
      { id: 106, label: '床垫', image: 'https://images.unsplash.com/photo-1505693416388-16fce068fe85?auto=format&fit=crop&w=360&q=80', productId: 103 },
      { id: 107, label: '家饰', image: 'https://images.unsplash.com/photo-1519710164239-da123dc03ef4?auto=format&fit=crop&w=360&q=80', productId: 101 },
      { id: 108, label: '家具', image: 'https://images.unsplash.com/photo-1505693416388-f0c5ce068fe85?auto=format&fit=crop&w=360&q=80', productId: 102 },
      { id: 109, label: '宠物', image: 'https://images.unsplash.com/photo-1517849845537-4d257902454a?auto=format&fit=crop&w=360&q=80', productId: 103 },
    ],
  },
  {
    id: 2,
    name: '餐厨',
    heroImage: 'https://images.unsplash.com/photo-1556911220-bff31c812dba?auto=format&fit=crop&w=960&q=80',
    items: [
      { id: 201, label: '锅具', image: 'https://images.unsplash.com/photo-1584990347449-a4c1a8b6d8f1?auto=format&fit=crop&w=360&q=80', productId: 101 },
      { id: 202, label: '杯壶', image: 'https://images.unsplash.com/photo-1514228742587-6b1558fcf93a?auto=format&fit=crop&w=360&q=80', productId: 102 },
      { id: 203, label: '餐具', image: 'https://images.unsplash.com/photo-1473093295043-cdd812d0e601?auto=format&fit=crop&w=360&q=80', productId: 103 },
      { id: 204, label: '烘焙', image: 'https://images.unsplash.com/photo-1517433670267-08bbd4be890f?auto=format&fit=crop&w=360&q=80', productId: 101 },
      { id: 205, label: '保鲜', image: 'https://images.unsplash.com/photo-1482049016688-2d3e1b311543?auto=format&fit=crop&w=360&q=80', productId: 102 },
      { id: 206, label: '收纳', image: 'https://images.unsplash.com/photo-1517705008128-361805f42e86?auto=format&fit=crop&w=360&q=80', productId: 103 },
    ],
  },
  {
    id: 3,
    name: '饮食',
    heroImage: 'https://images.unsplash.com/photo-1547592166-23ac45744acd?auto=format&fit=crop&w=960&q=80',
    items: [
      { id: 301, label: '零食', image: 'https://images.unsplash.com/photo-1515003197210-e0cd71810b5f?auto=format&fit=crop&w=360&q=80', productId: 101 },
      { id: 302, label: '茶饮', image: 'https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=360&q=80', productId: 102 },
      { id: 303, label: '冲调', image: 'https://images.unsplash.com/photo-1523906630133-f6934a1ab2b9?auto=format&fit=crop&w=360&q=80', productId: 103 },
      { id: 304, label: '粮油', image: 'https://images.unsplash.com/photo-1576186726115-4d51596775d1?auto=format&fit=crop&w=360&q=80', productId: 101 },
      { id: 305, label: '生鲜', image: 'https://images.unsplash.com/photo-1542838132-92c53300491e?auto=format&fit=crop&w=360&q=80', productId: 102 },
      { id: 306, label: '即食', image: 'https://images.unsplash.com/photo-1482049016688-2d3e1b311543?auto=format&fit=crop&w=360&q=80', productId: 103 },
    ],
  },
  {
    id: 4,
    name: '配件',
    heroImage: 'https://images.unsplash.com/photo-1521572267360-ee0c2909d518?auto=format&fit=crop&w=960&q=80',
    items: [
      { id: 401, label: '手机配件', image: 'https://images.unsplash.com/photo-1511707171634-5f897ff02aa9?auto=format&fit=crop&w=360&q=80', productId: 101 },
      { id: 402, label: '耳机', image: 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?auto=format&fit=crop&w=360&q=80', productId: 102 },
      { id: 403, label: '手表', image: 'https://images.unsplash.com/photo-1523275335684-37898b6baf30?auto=format&fit=crop&w=360&q=80', productId: 103 },
      { id: 404, label: '背包', image: 'https://images.unsplash.com/photo-1542291026-7eec264c27ff?auto=format&fit=crop&w=360&q=80', productId: 101 },
    ],
  },
  {
    id: 5,
    name: '服装',
    heroImage: 'https://images.unsplash.com/photo-1445205170230-053b83016050?auto=format&fit=crop&w=960&q=80',
    items: [
      { id: 501, label: '上装', image: 'https://images.unsplash.com/photo-1483985988355-763728e1935b?auto=format&fit=crop&w=360&q=80', productId: 101 },
      { id: 502, label: '下装', image: 'https://images.unsplash.com/photo-1473966968600-fa801b869a1a?auto=format&fit=crop&w=360&q=80', productId: 102 },
      { id: 503, label: '裙装', image: 'https://images.unsplash.com/photo-1496747611176-843222e1e57c?auto=format&fit=crop&w=360&q=80', productId: 103 },
      { id: 504, label: '内搭', image: 'https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?auto=format&fit=crop&w=360&q=80', productId: 101 },
    ],
  },
  {
    id: 6,
    name: '婴童',
    heroImage: 'https://images.unsplash.com/photo-1515488042361-ee00e0ddd4e4?auto=format&fit=crop&w=960&q=80',
    items: [
      { id: 601, label: '玩具', image: 'https://images.unsplash.com/photo-1516627145497-ae6968895b74?auto=format&fit=crop&w=360&q=80', productId: 101 },
      { id: 602, label: '童装', image: 'https://images.unsplash.com/photo-1519238359922-989348752efb?auto=format&fit=crop&w=360&q=80', productId: 102 },
      { id: 603, label: '出行', image: 'https://images.unsplash.com/photo-1519681393784-d120267933ba?auto=format&fit=crop&w=360&q=80', productId: 103 },
    ],
  },
  {
    id: 7,
    name: '杂货',
    heroImage: 'https://images.unsplash.com/photo-1517142089942-ba376ce32a2e?auto=format&fit=crop&w=960&q=80',
    items: [
      { id: 701, label: '收纳盒', image: 'https://images.unsplash.com/photo-1503602642458-232111445657?auto=format&fit=crop&w=360&q=80', productId: 101 },
      { id: 702, label: '文具', image: 'https://images.unsplash.com/photo-1455390582262-044cdead277a?auto=format&fit=crop&w=360&q=80', productId: 102 },
      { id: 703, label: '清洁工具', image: 'https://images.unsplash.com/photo-1583947582886-f40ec95dd752?auto=format&fit=crop&w=360&q=80', productId: 103 },
    ],
  },
  {
    id: 8,
    name: '洗护',
    heroImage: 'https://images.unsplash.com/photo-1526947425960-945c6e72858f?auto=format&fit=crop&w=960&q=80',
    items: [
      { id: 801, label: '洗发', image: 'https://images.unsplash.com/photo-1556228578-8c89e6adf883?auto=format&fit=crop&w=360&q=80', productId: 101 },
      { id: 802, label: '护肤', image: 'https://images.unsplash.com/photo-1522335789203-aabd1fc54bc9?auto=format&fit=crop&w=360&q=80', productId: 102 },
      { id: 803, label: '身体护理', image: 'https://images.unsplash.com/photo-1619451334792-150fd785ee74?auto=format&fit=crop&w=360&q=80', productId: 103 },
    ],
  },
  {
    id: 9,
    name: '志趣',
    heroImage: 'https://images.unsplash.com/photo-1516979187457-637abb4f9353?auto=format&fit=crop&w=960&q=80',
    items: [
      { id: 901, label: '露营', image: 'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=360&q=80', productId: 101 },
      { id: 902, label: '读物', image: 'https://images.unsplash.com/photo-1512820790803-83ca734da794?auto=format&fit=crop&w=360&q=80', productId: 102 },
      { id: 903, label: '香氛', image: 'https://images.unsplash.com/photo-1603006905003-be475563bc59?auto=format&fit=crop&w=360&q=80', productId: 103 },
    ],
  },
]

export const previewCartItems: CartItem[] = [
  {
    id: '1',
    sku_id: '1001',
    product_id: '103',
    product_title: '天然硅胶宠物除毛按摩刷',
    sku_image: 'https://images.unsplash.com/photo-1583511655857-d19b40a7a54e?auto=format&fit=crop&w=480&q=80',
    qty: 1,
    snapshot_price_cents: 3900,
    current_price_cents: 3900,
    available_stock: 18,
    sku_attrs: ['标准'],
    is_available: true,
  },
]

export const previewAddress = {
  name: 'jiechud',
  phone: '18311046191',
  detail: '123',
}

export const previewCouponSummary = {
  count: 0,
  title: '没有可用的优惠券',
}
