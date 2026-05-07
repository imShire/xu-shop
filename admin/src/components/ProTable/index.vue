<script setup lang="ts">
import { ref, computed } from 'vue'
import { RefreshRight, Setting } from '@element-plus/icons-vue'

export interface ColumnDef {
  prop?: string
  label: string
  width?: number | string
  minWidth?: number | string
  slot?: string
  formatter?: (row: any, value: any) => string
  hidden?: boolean
  align?: 'left' | 'center' | 'right'
  fixed?: 'left' | 'right' | boolean
}

interface Props {
  columns: ColumnDef[]
  data: any[]
  total?: number
  loading?: boolean
  rowKey?: string
  selection?: boolean
  pageSizes?: number[]
}

const props = withDefaults(defineProps<Props>(), {
  total: 0,
  loading: false,
  rowKey: 'id',
  selection: false,
  pageSizes: () => [10, 20, 50, 100],
})

const page = defineModel<number>('page', { default: 1 })
const pageSize = defineModel<number>('pageSize', { default: 20 })

const emit = defineEmits<{
  refresh: []
  selectionChange: [rows: any[]]
}>()

const hiddenCols = ref<Set<string>>(new Set())
const settingVisible = ref(false)

const visibleColumns = computed(() =>
  props.columns.filter((col) => !col.hidden && !hiddenCols.value.has(col.label))
)

function toggleCol(label: string) {
  if (hiddenCols.value.has(label)) {
    hiddenCols.value.delete(label)
  } else {
    hiddenCols.value.add(label)
  }
}
</script>

<template>
  <div class="pro-table">
    <div v-if="$slots.search" class="search-bar">
      <slot name="search" />
    </div>

    <div class="toolbar">
      <div class="toolbar-left">
        <slot name="toolbar" />
      </div>
      <div class="toolbar-right" style="display: flex; gap: 8px">
        <el-tooltip content="刷新">
          <el-button :icon="RefreshRight" circle size="small" @click="emit('refresh')" />
        </el-tooltip>
        <el-popover
          v-model:visible="settingVisible"
          placement="bottom-end"
          trigger="click"
          width="180"
        >
          <template #reference>
            <el-tooltip content="列设置">
              <el-button :icon="Setting" circle size="small" />
            </el-tooltip>
          </template>
          <div style="display: flex; flex-direction: column; gap: 6px">
            <el-checkbox
              v-for="col in columns"
              :key="col.label"
              :model-value="!hiddenCols.has(col.label)"
              :label="col.label"
              @change="toggleCol(col.label)"
            >
              {{ col.label }}
            </el-checkbox>
          </div>
        </el-popover>
      </div>
    </div>

    <el-table
      v-loading="loading"
      :data="data"
      :row-key="rowKey"
      border
      stripe
      style="width: 100%"
      @selection-change="(rows: any[]) => emit('selectionChange', rows)"
    >
      <el-table-column v-if="selection" type="selection" width="48" fixed="left" />

      <template v-for="col in visibleColumns" :key="col.prop || col.label">
        <el-table-column
          :prop="col.prop"
          :label="col.label"
          :width="col.width"
          :min-width="col.minWidth"
          :align="col.align || 'left'"
          :fixed="col.fixed"
          show-overflow-tooltip
        >
          <template v-if="col.slot || col.formatter" #default="{ row }">
            <slot v-if="col.slot" :name="col.slot" :row="row" />
            <span v-else-if="col.formatter">{{ col.formatter(row, col.prop ? row[col.prop] : undefined) }}</span>
          </template>
        </el-table-column>
      </template>

      <el-table-column v-if="$slots.actions" label="操作" fixed="right" width="160">
        <template #default="{ row }">
          <slot name="actions" :row="row" />
        </template>
      </el-table-column>
    </el-table>

    <div v-if="total > 0" class="pagination-wrapper">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="pageSizes"
        layout="total, sizes, prev, pager, next, jumper"
        background
      />
    </div>
  </div>
</template>
