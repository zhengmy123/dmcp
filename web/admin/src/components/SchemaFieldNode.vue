<template>
  <div class="field-node" :class="{ 'is-nested': depth > 0 }">
    <div class="field-row">
      <!-- 缩进线 -->
      <span v-if="depth > 0" class="indent-line" :style="{ left: `${(depth - 1) * 20}px` }"></span>

      <!-- 展开/折叠按钮 -->
      <button
        v-if="field.type === 'object' || field.type === 'array'"
        @click="toggleExpand"
        class="expand-btn"
        type="button"
      >
        {{ field.expanded ? '▼' : '▶' }}
      </button>
      <span v-else class="expand-placeholder"></span>

      <!-- 字段名 -->
      <input
        v-model="localName"
        @input="emitUpdate"
        class="field-name"
        placeholder="字段名"
      />

      <!-- 类型选择 -->
      <select v-model="localType" @change="onTypeChange" class="field-type">
        <option value="string">字符串</option>
        <option value="integer">整数</option>
        <option value="number">数字</option>
        <option value="boolean">布尔</option>
        <option value="array">数组</option>
        <option value="object">对象</option>
      </select>

      <!-- 描述 -->
      <input
        v-model="localDescription"
        @input="emitUpdate"
        class="field-desc"
        placeholder="描述"
      />

      <!-- 必填 -->
      <label class="field-required">
        <input v-model="localRequired" type="checkbox" @change="emitUpdate" />
        <span>必填</span>
      </label>

      <!-- 默认值 -->
      <input
        v-model="localDefault"
        @input="emitUpdate"
        class="field-default"
        placeholder="默认值"
        :disabled="isComplexType"
      />

      <!-- 删除按钮 -->
      <button @click="emitDelete" class="delete-btn" type="button" title="删除字段">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
        </svg>
      </button>
    </div>

    <!-- 子字段 -->
    <div v-if="isComplexType && field.expanded" class="children">
      <SchemaFieldNode
        v-for="child in field.children"
        :key="child.id"
        :field="child"
        :depth="depth + 1"
        @update="(id, updates) => emit('update', id, updates)"
        @delete="emit('delete', $event)"
      />
      <button @click="addChild" class="add-child-btn" type="button">
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
        </svg>
        添加子字段
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import type { SchemaField } from '@/utils/schemaBuilder';
import { createField } from '@/utils/schemaBuilder';

const props = defineProps<{
  field: SchemaField;
  depth: number;
}>();

const emit = defineEmits<{
  (e: 'update', id: string, updates: Partial<SchemaField>): void;
  (e: 'delete', id: string): void;
}>();

// 本地状态用于 v-model
const localName = ref(props.field.name);
const localType = ref(props.field.type);
const localDescription = ref(props.field.description || '');
const localRequired = ref(props.field.required || false);
const localDefault = ref(props.field.default ?? '');

// 同步外部 props 变化
watch(() => props.field, (newField) => {
  localName.value = newField.name;
  localType.value = newField.type;
  localDescription.value = newField.description || '';
  localRequired.value = newField.required || false;
  localDefault.value = newField.default ?? '';
}, { deep: true });

const isComplexType = computed(() => props.field.type === 'object' || props.field.type === 'array');

function toggleExpand() {
  emit('update', props.field.id, { expanded: !props.field.expanded });
}

function onTypeChange() {
  const updates: Partial<SchemaField> = {
    type: localType.value as SchemaField['type'],
  };
  // 如果切换到简单类型，清空 children
  if (!['object', 'array'].includes(localType.value)) {
    updates.children = [];
  } else if (!props.field.children || props.field.children.length === 0) {
    // 如果切换到复杂类型且没有 children，初始化空数组
    updates.children = [];
    updates.expanded = true;
  }
  emit('update', props.field.id, updates);
}

function emitUpdate() {
  emit('update', props.field.id, {
    name: localName.value,
    type: localType.value as SchemaField['type'],
    description: localDescription.value,
    required: localRequired.value,
    default: localDefault.value,
  });
}

function emitDelete() {
  emit('delete', props.field.id);
}

function addChild() {
  const newChild = createField();
  const children = [...(props.field.children || []), newChild];
  emit('update', props.field.id, { children, expanded: true });
}
</script>

<style scoped>
.field-node {
  position: relative;
  margin-bottom: 4px;
}

.field-node.is-nested {
  padding-left: 20px;
  border-left: 1px solid #e5e7eb;
  margin-left: 10px;
}

.indent-line {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 1px;
  background-color: #e5e7eb;
}

.field-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  background: #f9fafb;
  border-radius: 6px;
  flex-wrap: wrap;
}

.expand-btn {
  width: 20px;
  height: 20px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  cursor: pointer;
  color: #6b7280;
  font-size: 10px;
  border-radius: 4px;
}

.expand-btn:hover {
  background: #e5e7eb;
  color: #374151;
}

.expand-placeholder {
  width: 20px;
}

.field-name {
  width: 120px;
  padding: 4px 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 13px;
}

.field-type {
  width: 90px;
  padding: 4px 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 13px;
  background: white;
}

.field-desc {
  flex: 1;
  min-width: 100px;
  padding: 4px 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 13px;
}

.field-required {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #6b7280;
  white-space: nowrap;
}

.field-required input {
  width: 14px;
  height: 14px;
}

.field-default {
  width: 100px;
  padding: 4px 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 13px;
}

.field-default:disabled {
  background: #f3f4f6;
  cursor: not-allowed;
}

.delete-btn {
  width: 28px;
  height: 28px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  cursor: pointer;
  color: #9ca3af;
  border-radius: 4px;
}

.delete-btn:hover {
  background: #fee2e2;
  color: #dc2626;
}

.children {
  padding: 8px 0 8px 8px;
}

.add-child-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  margin-top: 4px;
  font-size: 12px;
  color: #4f46e5;
  background: transparent;
  border: 1px dashed #c7d2fe;
  border-radius: 4px;
  cursor: pointer;
}

.add-child-btn:hover {
  background: #eef2ff;
  border-color: #4f46e5;
}
</style>
