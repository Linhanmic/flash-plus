<template>
  <div class="flash-table">
    <el-table
      :data="store.treeData"
      style="width: 100%"
      row-key="id"
      :show-header="false"
      :expand-row-keys="store.expandedKeys"
      :tree-props="{ children: 'children' }"
      :indent="0"
      :row-class-name="rowClassName"
      empty-text="Waiting for specification events…"
      @expand-change="store.onExpandChange"
    >
      <el-table-column class-name="name-cell" min-width="320">
        <template #default="{ row }">
          <button
            v-if="hasChildren(row)"
            type="button"
            class="toggle"
            :class="{ expanded: isExpanded(row) }"
            :aria-label="isExpanded(row) ? 'collapse' : 'expand'"
            @click.stop="toggle(row)"
          />
          <span v-else class="toggle-spacer" />
          <div class="cell-content" :class="`type-${row.type}`">
            <div class="row-main">
              <span v-if="row.kind" class="kind">{{ row.kind }}</span>
              <span class="row-name" :title="row.fileName || row.name">{{ row.name }}</span>
              <span v-if="row.heading" class="row-heading">{{ row.heading }}</span>
            </div>
            <pre v-if="row.type === 'error' && row.stackTrace" class="stack">{{ row.stackTrace }}</pre>
          </div>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ElTable, ElTableColumn } from 'element-plus'
import { useReportStore } from '../stores/report'

const store = useReportStore()

function hasChildren(row) {
  return Array.isArray(row.children) && row.children.length > 0
}

function isExpanded(row) {
  return store.expandedKeys.includes(String(row.id))
}

function toggle(row) {
  store.onExpandChange(row, !isExpanded(row))
}

function rowClassName({ row }) {
  return [`is-${row.status || 'progress'}`, `is-${row.type}`, `is-level-${row.level ?? 0}`]
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
  --el-table-tr-bg-color: transparent;
  --el-table-row-hover-bg-color: transparent;
  --el-table-header-bg-color: transparent;
  background: transparent;
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Microsoft YaHei', Arial, sans-serif;
  font-size: 14px;
  color: #3d3d3d;
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
  border-spacing: 0 6px;
}

.flash-table :deep(.el-table td.el-table__cell) {
  border: none !important;
  background: transparent !important;
  padding: 0;
}

.flash-table :deep(.el-table__body tr:hover > td.el-table__cell) {
  background: transparent !important;
}

.flash-table :deep(.el-table .cell) {
  display: flex;
  flex-direction: row;
  align-items: center;
  width: 100%;
  box-sizing: border-box;
  padding: 8px 14px 8px 12px;
  line-height: 20px;
  overflow: visible;
  background: #fff;
  border: 1px solid #e5e5e5;
  border-left: 4px solid #1abc9c;
}

.flash-table :deep(.el-table__row.is-level-0 td.el-table__cell) {
  padding-left: 0;
}

.flash-table :deep(.el-table__row.is-level-1 td.el-table__cell) {
  padding-left: 36px;
}

.flash-table :deep(.el-table__row.is-level-2 td.el-table__cell) {
  padding-left: 72px;
}

.flash-table :deep(.el-table__row.is-level-3 td.el-table__cell) {
  padding-left: 108px;
}

.flash-table :deep(.el-table__row.is-level-4 td.el-table__cell) {
  padding-left: 144px;
}

.flash-table :deep(.el-table__indent),
.flash-table :deep(.el-table__placeholder),
.flash-table :deep(.el-table__expand-icon) {
  display: none !important;
}

.toggle,
.toggle-spacer {
  width: 16px;
  height: 16px;
  margin-right: 8px;
  flex-shrink: 0;
}

.toggle {
  border: 0;
  padding: 0;
  border-radius: 3px;
  background: #3a3a3a;
  cursor: pointer;
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.toggle::before {
  content: '';
  width: 8px;
  height: 2px;
  background: #fff;
  border-radius: 1px;
}

.toggle:not(.expanded)::after {
  content: '';
  position: absolute;
  width: 2px;
  height: 8px;
  background: #fff;
  border-radius: 1px;
}

.flash-table :deep(.el-table__row.is-fail .cell) {
  border-left-color: #e74c3c;
}

.flash-table :deep(.el-table__row.is-progress .cell) {
  border-left-color: #95a5a6;
}

.flash-table :deep(.el-table__row.is-skip .cell) {
  border-left-color: #bdc3c7;
}

.flash-table :deep(.el-table__row.is-error .cell) {
  border-left-color: #e74c3c;
  background: #fff8f8;
}

.flash-table :deep(.el-table__row:hover .cell) {
  background: #f7fbfa;
}

.flash-table :deep(.el-table__row.is-error:hover .cell) {
  background: #fff1f1;
}

.cell-content {
  display: flex;
  flex-direction: column;
  justify-content: center;
  flex: 1;
  min-width: 0;
}

.row-main {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.kind {
  flex-shrink: 0;
  font-size: 11px;
  line-height: 18px;
  padding: 0 6px;
  border-radius: 3px;
  background: #f2f2f2;
  color: #888;
}

.type-spec .kind {
  background: #e8f8f5;
  color: #148f77;
}

.type-scenario .kind {
  background: #eef3f7;
  color: #5d6d7e;
}

.type-concept .kind {
  background: #f5eef8;
  color: #8e44ad;
}

.type-step .kind {
  background: #f4f4f4;
  color: #7f8c8d;
}

.row-name {
  line-height: 20px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.type-spec .row-name {
  font-weight: 600;
  font-size: 14px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}

.type-scenario .row-name {
  font-weight: 500;
}

.type-concept .row-name {
  font-weight: 500;
}

.row-heading {
  color: #8a8a8a;
  font-size: 12px;
  font-weight: 400;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.type-error .row-name {
  color: #c0392b;
  font-size: 13px;
  white-space: pre-wrap;
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
