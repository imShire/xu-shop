import fs from 'node:fs'
import path from 'node:path'

const root = process.cwd()
const sourceDir = path.join(root, 'docs', 'address')
const outputFile = path.join(root, 'server', 'cmd', 'cli', 'assets', 'regions.json')

function readJson(file) {
  return JSON.parse(fs.readFileSync(path.join(sourceDir, file), 'utf8'))
}

const provinces = readJson('province.json')
const cities = readJson('city.json')
const areas = readJson('area.json')
const streets = readJson('street.json')

const rows = []
const seen = new Set()

function appendRows(nodes, parentCode, level) {
  nodes.forEach((node, index) => {
    const code = String(node.id)
    if (seen.has(code)) return
    seen.add(code)
    rows.push({
      code,
      parent_code: parentCode,
      name: node.name,
      level,
      sort: index + 1,
    })
  })
}

appendRows(provinces, '', 1)
Object.entries(cities).forEach(([parentCode, nodes]) => appendRows(nodes, parentCode, 2))
Object.entries(areas).forEach(([parentCode, nodes]) => appendRows(nodes, parentCode, 3))
Object.entries(streets).forEach(([parentCode, nodes]) => appendRows(nodes, parentCode, 4))

fs.mkdirSync(path.dirname(outputFile), { recursive: true })
fs.writeFileSync(outputFile, `${JSON.stringify(rows, null, 2)}\n`)

console.log(`wrote ${rows.length} rows to ${outputFile}`)
