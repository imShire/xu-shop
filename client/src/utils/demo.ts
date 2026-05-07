import type { CartItem, Category, Product, TimelineItem, User } from '@/types/biz'

// Placeholder seeds keep the shell readable before backend modules are populated in local dev.
export const demoCategories: Category[] = [
  { id: '1', name: '春季新作', tagline: '茶香、亚麻与轻裙' },
  { id: '2', name: '餐桌补给', tagline: '囤一点耐看的日用' },
  { id: '3', name: '私域好评', tagline: '复购率最高的老朋友' },
]

export const demoProducts: Product[] = [
  {
    id: '101',
    title: '雨前龙井礼盒',
    subtitle: '清润不苦，适合送礼与自留',
    main_image: 'https://images.unsplash.com/photo-1547826039-bfc35e0f1ea8?auto=format&fit=crop&w=900&q=80',
    price_min_cents: 16800,
    price_max_cents: 16800,
    tags: ['48h 发货', '茶礼'],
    category_id: '1',
  },
  {
    id: '102',
    title: '手作香氛蜡烛',
    subtitle: '木质雪松调，适合夜读',
    main_image: 'https://images.unsplash.com/photo-1603006905003-be475563bc59?auto=format&fit=crop&w=900&q=80',
    price_min_cents: 9800,
    price_max_cents: 9800,
    tags: ['限量', '礼物'],
    category_id: '1',
  },
  {
    id: '103',
    title: '亚麻桌布套装',
    subtitle: '粗纺手感，适合餐桌换季',
    main_image: 'https://images.unsplash.com/photo-1513694203232-719a280e022f?auto=format&fit=crop&w=900&q=80',
    price_min_cents: 23800,
    price_max_cents: 23800,
    tags: ['家居', '复购高'],
    category_id: '2',
  },
]

// Placeholder cart rows keep the cart shell understandable before Phase 3 APIs are wired end to end.
export const demoCartItems: CartItem[] = [
  {
    id: '901',
    sku_id: '9011',
    product_id: '101',
    product_title: '雨前龙井礼盒',
    sku_image: demoProducts[0].main_image,
    qty: 1,
    snapshot_price_cents: 16800,
    current_price_cents: 16800,
    available_stock: 12,
    sku_attrs: ['单盒装'],
    is_available: true,
  },
  {
    id: '902',
    sku_id: '9021',
    product_id: '103',
    product_title: '亚麻桌布套装',
    sku_image: demoProducts[2].main_image,
    qty: 1,
    snapshot_price_cents: 23800,
    current_price_cents: 23800,
    available_stock: 6,
    sku_attrs: ['奶杏色'],
    is_available: true,
  },
]

export const demoUser: User = {
  id: '7',
  phone: null,
  nickname: '私域访客',
  avatar: null,
  gender: 0,
  status: 'active',
}

export const demoTimeline: TimelineItem[] = [
  { id: '1', title: '浏览过 雨前龙井礼盒', summary: '最近一次停留 4 分钟', time: '今天 18:20' },
  { id: '2', title: '收藏了 亚麻桌布套装', summary: '适合换季餐桌布置', time: '昨天 21:04' },
  { id: '3', title: '查看了 香氛蜡烛', summary: '晚间安静调性', time: '昨天 20:12' },
]
