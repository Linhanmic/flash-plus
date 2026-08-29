<template>
  <div class="flash-table">
    <el-table
      :data="store.treeData"
      style="width: 100%"
      row-key="id"
      :show-header="false"
      :expand-row-keys="expandedKeys"
      :tree-props="{ children: 'children' }"
      :indent="24"
      :row-class-name="rowClassName"
      empty-text="Waiting for specification events…"
      @expand-change="onExpandChange"
    >
      <el-table-column class-name="name-cell" min-width="320">
        <template #default="{ row }">
          <div class="cell-content" :class="`type-${row.type}`">
            <span class="row-name">{{ row.name }}</span>
            <pre v-if="row.type === 'error' && row.stackTrace" class="stack">{{ row.stackTrace }}</pre>
          </div>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { ElTable, ElTableColumn } from 'element-plus'
import { useEventStore } from '../stores/eventStore'

const props = defineProps({
  allExpanded: {
    type: Boolean,
    default: true
  }
})

const store = useEventStore()
const expandedKeys = ref([])

function collectExpandableIds(nodes, acc = []) {
  for (const node of nodes || []) {
    if (node.children?.length) {
      acc.push(String(node.id))
      collectExpandableIds(node.children, acc)
    }
  }
  return acc
}

function syncExpanded() {
  if (props.allExpanded) {
    expandedKeys.value = collectExpandableIds(store.treeData)
  } else {
    expandedKeys.value = []
  }
}

watch(
  () => [props.allExpanded, store.treeData],
  () => syncExpanded(),
  { immediate: true, deep: true }
)

function onExpandChange(row, expanded) {
  const id = String(row.id)
  const open = typeof expanded === 'boolean' ? expanded : expandedKeys.value.includes(id)
  if (typeof expanded !== 'boolean') {
    return
  }
  if (open) {
    if (!expandedKeys.value.includes(id)) {
      expandedKeys.value = [...expandedKeys.value, id]
    }
  } else {
    expandedKeys.value = expandedKeys.value.filter((key) => key !== id)
  }
}

function rowClassName({ row }) {
  return [`is-${row.status || 'progress'}`, `is-${row.type}`]
}
</script>

<style scoped>
.flash-table {
  margin-top: 12px;
  background: transparent;
}

.flash-table :deep(.el-table) {
  --el-table-border-color: #e6e6e6;
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: #ffffff;
  --el-table-row-hover-bg-color: #f7fbfa;
  --el-table-header-bg-color: transparent;
  background: transparent;
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Microsoft YaHei', Arial, sans-serif;
  font-size: 14px;
  color: #4a4a4a;
}

.flash-table :deep(.el-table::before),
.flash-table :deep(.el-table__inner-wrapper::before) {
  display: none;
}

.flash-table :deep(.el-table__inner-wrapper) {
  background: transparent;
}

.flash-table :deep(.el-table__body) {
  border-collapse: separate;
  border-spacing: 0 8px;
}

.flash-table :deep(.el-table td.el-table__cell) {
  border: 1px solid #e5e5e5;
  border-left: 4px solid #1abc9c;
  background: #fff;
  padding: 10px 14px;
}

.flash-table :deep(.el-table__row.is-fail td.el-table__cell) {
  border-left-color: #e74c3c;
}

.flash-table :deep(.el-table__row.is-progress td.el-table__cell) {
  border-left-color: #95a5a6;
}

.flash-table :deep(.el-table__row.is-skip td.el-table__cell) {
  border-left-color: #bdc3c7;
}

.flash-table :deep(.el-table__row.is-error td.el-table__cell) {
  border-left-color: #e74c3c;
  background: #fff8f8;
}

.flash-table :deep(.el-table__expand-icon) {
  width: 16px;
  height: 16px;
  margin-right: 10px;
  border-radius: 3px;
  background: #3a3a3a;
  color: #fff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transform: none !important;
  position: relative;
  vertical-align: middle;
}

.flash-table :deep(.el-table__expand-icon .el-icon) {
  display: none;
}

.flash-table :deep(.el-table__expand-icon::before) {
  content: '';
  width: 8px;
  height: 2px;
  background: #fff;
  border-radius: 1px;
}

.flash-table :deep(.el-table__expand-icon:not(.el-table__expand-icon--expanded)::after) {
  content: '';
  position: absolute;
  width: 2px;
  height: 8px;
  background: #fff;
  border-radius: 1px;
}

.flash-table :deep(.el-table__placeholder) {
  width: 16px;
  height: 16px;
  margin-right: 10px;
}

.cell-content {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-height: 20px;
}

.row-name {
  line-height: 20px;
}

.type-spec .row-name {
  font-weight: 600;
}

.type-error .row-name {
  color: #c0392b;
  font-size: 13px;
}

.stack {
  margin: 6px 0 0;
  color: #8a3030;
  font-size: 12px;
  line-height: 16px;
  white-space: pre-wrap;
  font-family: inherit;
}
</style>
