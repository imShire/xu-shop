#!/usr/bin/env python3
"""抓取 CRMEB 演示站数据，下载图片到本地，生成 seed.sql"""

import json, os, re, sys, time, urllib.request, urllib.parse
from datetime import datetime, timezone

BASE_API   = "https://api.front.merchant.java.crmeb.net"
UA         = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15"
REFERER    = "https://h5.merchant.java.crmeb.net/"
UPLOAD_DIR = os.path.join(os.path.dirname(__file__), "../server/uploads/seed")
LOCAL_BASE = "http://localhost:8080/uploads/seed"
SQL_OUT    = os.path.join(os.path.dirname(__file__), "../seed.sql")

# ── helpers ──────────────────────────────────────────────────────────────────

def fetch(path, params=None):
    url = f"{BASE_API}{path}"
    if params:
        url += "?" + urllib.parse.urlencode({k: v for k, v in params.items() if v is not None})
    req = urllib.request.Request(url, headers={"User-Agent": UA, "Referer": REFERER})
    with urllib.request.urlopen(req, timeout=15) as r:
        return json.loads(r.read())

def download(url, folder, name):
    """下载图片，返回本地 URL；失败返回原始 URL"""
    if not url:
        return None
    ext = (re.search(r'\.(png|jpg|jpeg|webp|gif)', url, re.I) or [None, 'jpg'])[1].lower()
    safe = re.sub(r'[^a-z0-9_-]', '_', name.lower())[:60]
    dest_dir = os.path.join(UPLOAD_DIR, folder)
    os.makedirs(dest_dir, exist_ok=True)
    filename = f"{safe}.{ext}"
    dest = os.path.join(dest_dir, filename)
    if not os.path.exists(dest):
        try:
            req = urllib.request.Request(url, headers={"User-Agent": UA, "Referer": REFERER})
            with urllib.request.urlopen(req, timeout=20) as r:
                data = r.read()
            with open(dest, "wb") as f:
                f.write(data)
            print(f"  ✓ {folder}/{filename} ({len(data)//1024}KB)")
        except Exception as e:
            print(f"  ✗ {name}: {e}", file=sys.stderr)
            return url   # 降级：用原始 URL
    return f"{LOCAL_BASE}/{folder}/{filename}"

def esc(s):
    if s is None:
        return "NULL"
    return "'" + str(s).replace("'", "''")[:500] + "'"

def cents(price_str):
    try:
        return int(round(float(price_str) * 100))
    except:
        return 0

now_sql = "NOW()"

# ── 1. 抓 Banner ────────────────────────────────────────────────────────────

print("\n📥 抓取 Banner...")
idx_data = fetch("/api/front/index/info")["data"]
raw_banners = idx_data.get("banner", [])
banners = []
for i, b in enumerate(raw_banners):
    local_img = download(b.get("pic"), "banners", f"banner_{i+1}")
    banners.append({
        "id":        i + 1,
        "title":     b.get("name", ""),
        "image_url": local_img or "",
        "link_url":  b.get("url", ""),
        "sort":      i,
        "is_active": True,
    })
print(f"  → {len(banners)} 条 banner")

# ── 2. 抓分类树 ─────────────────────────────────────────────────────────────

print("\n📥 抓取分类树...")
cat_data = fetch("/api/front/product/category/get/first")["data"]

flat_cats   = []  # [{local_id, pid_local, name, icon_url, sort, level}]
crmeb_to_local = {}  # crmeb_id -> local_id

def flatten(nodes, parent_local_id=0):
    for node in nodes:
        if not node.get("isShow", True):
            continue
        local_id = len(flat_cats) + 1
        crmeb_to_local[node["id"]] = local_id
        icon = download(node.get("icon"), "categories", f"cat_{node['id']}_{node['name']}")
        flat_cats.append({
            "local_id":  local_id,
            "parent_id": parent_local_id,
            "name":      node["name"],
            "icon":      icon or "",
            "sort":      node.get("sort", 0),
            "level":     node.get("level", 1),
        })
        if node.get("childList"):
            flatten(node["childList"], local_id)

flatten(cat_data)
print(f"  → {len(flat_cats)} 个分类节点")

# ── 3. 抓商品（全量） ────────────────────────────────────────────────────────

print("\n📥 抓取商品列表...")
prod_resp = fetch("/api/front/product/list", {"page": 1, "limit": 100})
raw_prods = prod_resp["data"]["list"]
print(f"  → 共 {len(raw_prods)} 件商品，开始下载图片...")

products = []
for i, p in enumerate(raw_prods):
    img = download(p.get("image"), "products", f"prod_{p['id']}")
    # 映射分类：优先使用 crmeb categoryId，否则用第一个叶子分类
    cat_local = crmeb_to_local.get(p.get("categoryId"))
    if not cat_local:
        # 找第一个叶子分类（level=3 或无子节点）
        for c in flat_cats:
            if c["level"] == 3 or c["level"] is None:
                cat_local = c["local_id"]
                break
        if not cat_local:
            cat_local = flat_cats[-1]["local_id"] if flat_cats else 1

    price_c = cents(p.get("price", "0"))
    ot_price_c = cents(p.get("otPrice", "0"))
    products.append({
        "id":              i + 1,
        "category_id":     cat_local,
        "title":           (p.get("name") or "")[:60],
        "subtitle":        f"来自 {p.get('merName','')} · 已售{p.get('sales',0)}件"[:120],
        "main_image":      img or p.get("image", ""),
        "status":          "onsale",
        "sales":           p.get("sales", 0),
        "virtual_sales":   p.get("ficti", 0),
        "price_min_cents": price_c,
        "price_max_cents": ot_price_c if ot_price_c > price_c else price_c,
        "unit":            p.get("unitName", "件"),
        "stock":           p.get("stock", 0),
    })
    if (i + 1) % 10 == 0:
        print(f"  ... {i+1}/{len(raw_prods)}")

print(f"  → {len(products)} 件商品处理完毕")

# ── 4. 生成 SQL ───────────────────────────────────────────────────────────────

print(f"\n📝 生成 {SQL_OUT} ...")

lines = [
    "-- ============================================================",
    "-- xu-shop 种子数据（由 scripts/seed_from_crmeb.py 生成）",
    f"-- 生成时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}",
    "-- 执行方式: psql $DB_DSN -f seed.sql",
    "-- ============================================================",
    "",
    "BEGIN;",
    "",
    "-- 清空（按外键顺序）",
    "DELETE FROM sku;",
    "DELETE FROM product;",
    "DELETE FROM category;",
    "DELETE FROM banner;",
    "",
]

# Banner
lines += [
    "-- ── Banner ──────────────────────────────────────────────────",
    "INSERT INTO banner (id, title, image_url, link_url, sort, is_active, created_at, updated_at) VALUES",
]
banner_rows = []
for b in banners:
    banner_rows.append(
        f"  ({b['id']}, {esc(b['title'])}, {esc(b['image_url'])}, {esc(b['link_url'])}, "
        f"{b['sort']}, {str(b['is_active']).lower()}, NOW(), NOW())"
    )
lines.append(",\n".join(banner_rows) + ";")
lines.append("")

# Category
lines += [
    "-- ── Category ────────────────────────────────────────────────",
    "INSERT INTO category (id, parent_id, name, icon, sort, status, created_at, updated_at) VALUES",
]
cat_rows = []
for c in flat_cats:
    cat_rows.append(
        f"  ({c['local_id']}, {c['parent_id']}, {esc(c['name'])}, "
        f"{esc(c['icon'])}, {c['sort']}, 'enabled', NOW(), NOW())"
    )
lines.append(",\n".join(cat_rows) + ";")
lines.append("")

# Product
lines += [
    "-- ── Product ─────────────────────────────────────────────────",
    "INSERT INTO product (id, category_id, title, subtitle, main_image, images, status,",
    "                     sales, virtual_sales, price_min_cents, price_max_cents, unit,",
    "                     is_virtual, tags, created_at, updated_at) VALUES",
]
prod_rows = []
for p in products:
    img_json = json.dumps([p["main_image"]]) if p["main_image"] else "[]"
    prod_rows.append(
        f"  ({p['id']}, {p['category_id']}, {esc(p['title'])}, {esc(p['subtitle'])},\n"
        f"   {esc(p['main_image'])}, {esc(img_json)}::jsonb, 'onsale',\n"
        f"   {p['sales']}, {p['virtual_sales']}, {p['price_min_cents']}, {p['price_max_cents']},\n"
        f"   {esc(p['unit'])}, false, '[]'::jsonb, NOW(), NOW())"
    )
lines.append(",\n".join(prod_rows) + ";")
lines.append("")

# SKU（每个商品一个默认 SKU）
lines += [
    "-- ── SKU（每商品一个默认规格） ───────────────────────────────",
    "INSERT INTO sku (id, product_id, attrs, price_cents, original_price_cents, stock, locked_stock, status, created_at, updated_at) VALUES",
]
sku_rows = []
for p in products:
    sku_rows.append(
        f"  ({p['id']}, {p['id']}, '{{}}'::jsonb, {p['price_min_cents']}, "
        f"{p['price_max_cents']}, {p['stock']}, 0, 'active', NOW(), NOW())"
    )
lines.append(",\n".join(sku_rows) + ";")
lines.append("")
lines.append("COMMIT;")
lines.append("")
lines.append("-- 验证")
lines.append("SELECT 'banner' AS tbl, COUNT(*) FROM banner")
lines.append("UNION ALL SELECT 'category', COUNT(*) FROM category")
lines.append("UNION ALL SELECT 'product',  COUNT(*) FROM product")
lines.append("UNION ALL SELECT 'sku',      COUNT(*) FROM sku;")

with open(SQL_OUT, "w", encoding="utf-8") as f:
    f.write("\n".join(lines))

print(f"✅ 完成！")
print(f"   Banner  : {len(banners)} 条")
print(f"   Category: {len(flat_cats)} 个")
print(f"   Product : {len(products)} 件")
print(f"   SKU     : {len(products)} 个（每商品 1 个默认）")
print(f"\n执行：")
print(f"   psql postgres://shop:shop@localhost:5432/shop -f seed.sql")
