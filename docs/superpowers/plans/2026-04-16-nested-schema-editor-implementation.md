# 多层嵌套 Schema 编辑器实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 HTTP 服务管理页面添加多层嵌套 Schema 可视化编辑功能，支持手动构建和自动从 JSON 样本生成。

**Architecture:** 使用递归树形结构存储嵌套字段，前端通过递归组件渲染，支持三种编辑模式（可视化、JSON、粘贴解析），与现有 ServicesPage.vue 集成。

**Tech Stack:** Vue 3 (Composition API), TypeScript

---

## 文件结构

```
web/admin/src/
├── components/
│   └── SchemaFieldNode.vue          # 递归字段节点组件 (新建)
├── utils/
│   └── schemaBuilder.ts             # 核心转换函数 (新建)
└── pages/
    └── ServicesPage.vue             # 修改：替换字段编辑逻辑 (修改)
```

---

## 实现任务

### Task 1: 创建 schemaBuilder.ts 工具函数

**Files:**
- Create: `web/admin/src/utils/schemaBuilder.ts`

- [ ] **Step 1: 创建 schemaBuilder.ts 文件**

```typescript
// SchemaField 类型定义
export interface SchemaField {
  id: string;
  name: string;
  type: 'string' | 'integer' | 'number' | 'boolean' | 'array' | 'object';
  description?: string;
  required?: boolean;
  default?: any;
  children?: SchemaField[];
  expanded?: boolean;
}

// 生成唯一 ID
export function generateFieldId(prefix = ''): string {
  return prefix ? `${prefix}-${Date.now()}-${Math.random().toString(36).substr(2, 6)}` : `field-${Date.now()}-${Math.random().toString(36).substr(2, 6)}`;
}

// 推断 JSON 值的类型
export function inferFieldType(value: any): SchemaField['type'] {
  if (value === null) return 'string';
  if (typeof value === 'string') return 'string';
  if (typeof value === 'number') return Number.isInteger(value) ? 'integer' : 'number';
  if (typeof value === 'boolean') return 'boolean';
  if (Array.isArray(value)) return 'array';
  if (typeof value === 'object') return 'object';
  return 'string';
}

// 解析默认值
export function parseDefaultValue(value: any, type: SchemaField['type']): any {
  if (value === undefined || value === null) return undefined;
  switch (type) {
    case 'integer': return parseInt(String(value), 10) || 0;
    case 'number': return parseFloat(String(value)) || 0;
    case 'boolean': return Boolean(value);
    default: return String(value);
  }
}

// 格式化默认值用于显示
export function formatDefaultValue(value: any): string {
  if (value === undefined || value === null) return '';
  if (typeof value === 'object') return JSON.stringify(value);
  return String(value);
}

// fieldsToSchema: 嵌套字段树 → JSON Schema (递归)
export function fieldsToSchema(fields: SchemaField[]): any {
  if (!fields || fields.length === 0) {
    return { type: 'object', properties: {} };
  }

  const properties: Record<string, any> = {};
  const required: string[] = [];

  for (const field of fields) {
    if (!field.name) continue;

    let propSchema: any;

    switch (field.type) {
      case 'object':
        propSchema = {
          type: 'object',
          properties: fieldsToSchema(field.children || []),
        };
        break;
      case 'array':
        propSchema = {
          type: 'array',
          items: {
            type: 'object',
            properties: fieldsToSchema(field.children || []),
          },
        };
        break;
      default:
        propSchema = { type: field.type };
    }

    if (field.description) {
      propSchema.description = field.description;
    }
    if (field.default !== undefined && field.default !== '') {
      propSchema.default = parseDefaultValue(field.default, field.type);
    }

    properties[field.name] = propSchema;

    if (field.required) {
      required.push(field.name);
    }
  }

  return {
    type: 'object',
    properties,
    ...(required.length > 0 ? { required } : {}),
  };
}

// schemaToFields: JSON Schema → 嵌套字段树 (递归)
export function schemaToFields(schema: any, parentId = ''): SchemaField[] {
  if (!schema || !schema.properties) {
    return [];
  }

  const required = schema.required || [];
  const fields: SchemaField[] = [];

  for (const [name, prop] of Object.entries(schema.properties)) {
    const id = parentId ? `${parentId}-${name}` : name;
    const propObj = prop as any;

    let type = propObj.type || 'string';
    let children: SchemaField[] | undefined;

    // 处理 array 的 items
    if (type === 'array' && propObj.items) {
      const itemsType = propObj.items.type;
      if (itemsType === 'object' || !itemsType) {
        children = schemaToFields(propObj.items, id);
        type = 'array';
      }
    }

    // 处理 object
    if (type === 'object' || (propObj.properties && !propObj.items)) {
      children = schemaToFields(propObj, id);
      type = 'object';
    }

    fields.push({
      id,
      name,
      type,
      description: propObj.description || '',
      required: required.includes(name),
      default: formatDefaultValue(propObj.default),
      children,
      expanded: true,
    });
  }

  return fields;
}

// parseJSONToFields: 样本 JSON → 嵌套字段树 (递归)
export function parseJSONToFields(obj: any, parentId = ''): SchemaField[] {
  if (obj === null || obj === undefined) {
    return [];
  }

  if (Array.isArray(obj)) {
    if (obj.length > 0) {
      const itemsId = parentId ? `${parentId}-items` : 'items';
      return [{
        id: itemsId,
        name: 'items',
        type: 'array',
        children: parseJSONToFields(obj[0], itemsId),
        expanded: true,
      }];
    }
    return [];
  }

  if (typeof obj !== 'object') {
    return [];
  }

  const fields: SchemaField[] = [];
  for (const [key, value] of Object.entries(obj)) {
    const id = parentId ? `${parentId}-${key}` : key;
    let type = inferFieldType(value);
    let children: SchemaField[] | undefined;

    if ((type === 'object' || type === 'array') && typeof value === 'object') {
      children = parseJSONToFields(value, id);
    }

    fields.push({
      id,
      name: key,
      type,
      description: '',
      required: false,
      default: '',
      children,
      expanded: true,
    });
  }

  return fields;
}

// 创建新字段
export function createField(partial?: Partial<SchemaField>): SchemaField {
  return {
    id: generateFieldId(),
    name: '',
    type: 'string',
    description: '',
    required: false,
    default: '',
    children: [],
    expanded: true,
    ...partial,
  };
}

// 递归查找字段
export function findFieldById(fields: SchemaField[], id: string): SchemaField | null {
  for (const field of fields) {
    if (field.id === id) return field;
    if (field.children) {
      const found = findFieldById(field.children, id);
      if (found) return found;
    }
  }
  return null;
}

// 递归删除字段
export function removeFieldById(fields: SchemaField[], id: string): SchemaField[] {
  return fields.filter(field => {
    if (field.id === id) return false;
    if (field.children) {
      field.children = removeFieldById(field.children, id);
    }
    return true;
  });
}

// 递归更新字段
export function updateFieldById(fields: SchemaField[], id: string, updates: Partial<SchemaField>): SchemaField[] {
  return fields.map(field => {
    if (field.id === id) {
      return { ...field, ...updates };
    }
    if (field.children) {
      return { ...field, children: updateFieldById(field.children, id, updates) };
    }
    return field;
  });
}
```

- [ ] **Step 2: 创建测试文件 schemaBuilder.test.ts**

```typescript
import {
  fieldsToSchema,
  schemaToFields,
  parseJSONToFields,
  createField,
  removeFieldById,
  updateFieldById,
  findFieldById,
} from './schemaBuilder';

describe('schemaBuilder', () => {
  describe('fieldsToSchema', () => {
    it('should convert empty fields to default schema', () => {
      const result = fieldsToSchema([]);
      expect(result).toEqual({ type: 'object', properties: {} });
    });

    it('should convert flat fields to schema', () => {
      const fields = [
        { id: '1', name: 'code', type: 'integer' as const, required: true, description: '', default: '' },
        { id: '2', name: 'message', type: 'string' as const, required: true, description: 'message desc', default: '' },
      ];
      const result = fieldsToSchema(fields);
      expect(result).toEqual({
        type: 'object',
        properties: {
          code: { type: 'integer' },
          message: { type: 'string', description: 'message desc' },
        },
        required: ['code', 'message'],
      });
    });

    it('should convert nested fields to schema', () => {
      const fields = [
        {
          id: '1', name: 'data', type: 'object' as const, required: false, description: '',
          children: [
            { id: '1-1', name: 'error', type: 'string' as const, required: false, description: '', default: '' },
            { id: '1-2', name: 'stdout', type: 'string' as const, required: false, description: '', default: '' },
          ],
        },
      ];
      const result = fieldsToSchema(fields);
      expect(result).toEqual({
        type: 'object',
        properties: {
          data: {
            type: 'object',
            properties: {
              error: { type: 'string' },
              stdout: { type: 'string' },
            },
          },
        },
      });
    });

    it('should convert array fields to schema', () => {
      const fields = [
        {
          id: '1', name: 'items', type: 'array' as const, required: false, description: '',
          children: [
            { id: '1-1', name: 'name', type: 'string' as const, required: false, description: '', default: '' },
          ],
        },
      ];
      const result = fieldsToSchema(fields);
      expect(result).toEqual({
        type: 'object',
        properties: {
          items: {
            type: 'array',
            items: {
              type: 'object',
              properties: {
                name: { type: 'string' },
              },
            },
          },
        },
      });
    });
  });

  describe('schemaToFields', () => {
    it('should convert empty schema to empty fields', () => {
      const result = schemaToFields({ type: 'object', properties: {} });
      expect(result).toEqual([]);
    });

    it('should convert flat schema to fields', () => {
      const schema = {
        type: 'object',
        properties: {
          code: { type: 'integer' },
          message: { type: 'string', description: 'msg desc' },
        },
        required: ['code'],
      };
      const result = schemaToFields(schema);
      expect(result).toHaveLength(2);
      expect(result[0]).toMatchObject({ name: 'code', type: 'integer', required: true });
      expect(result[1]).toMatchObject({ name: 'message', type: 'string', required: false });
    });

    it('should convert nested schema to fields', () => {
      const schema = {
        type: 'object',
        properties: {
          data: {
            type: 'object',
            properties: {
              error: { type: 'string' },
            },
          },
        },
      };
      const result = schemaToFields(schema);
      expect(result).toHaveLength(1);
      expect(result[0]).toMatchObject({ name: 'data', type: 'object' });
      expect(result[0].children).toHaveLength(1);
      expect(result[0].children?.[0]).toMatchObject({ name: 'error', type: 'string' });
    });

    it('should convert array schema to fields', () => {
      const schema = {
        type: 'object',
        properties: {
          items: {
            type: 'array',
            items: {
              type: 'object',
              properties: {
                name: { type: 'string' },
              },
            },
          },
        },
      };
      const result = schemaToFields(schema);
      expect(result).toHaveLength(1);
      expect(result[0]).toMatchObject({ name: 'items', type: 'array' });
      expect(result[0].children).toHaveLength(1);
    });
  });

  describe('parseJSONToFields', () => {
    it('should parse simple JSON object', () => {
      const json = { code: 0, message: 'success' };
      const result = parseJSONToFields(json);
      expect(result).toHaveLength(2);
      expect(result[0]).toMatchObject({ name: 'code', type: 'integer' });
      expect(result[1]).toMatchObject({ name: 'message', type: 'string' });
    });

    it('should parse nested JSON object', () => {
      const json = {
        code: 0,
        data: {
          error: '',
          status_code: 200,
        },
      };
      const result = parseJSONToFields(json);
      expect(result).toHaveLength(2);
      expect(result[0]).toMatchObject({ name: 'code', type: 'integer' });
      expect(result[1]).toMatchObject({ name: 'data', type: 'object' });
      expect(result[1].children).toHaveLength(2);
    });

    it('should infer array type', () => {
      const json = { items: ['a', 'b', 'c'] };
      const result = parseJSONToFields(json);
      expect(result[0]).toMatchObject({ name: 'items', type: 'array' });
    });

    it('should handle nested array with objects', () => {
      const json = { users: [{ name: 'Alice' }, { name: 'Bob' }] };
      const result = parseJSONToFields(json);
      expect(result[0]).toMatchObject({ name: 'users', type: 'array' });
      expect(result[0].children?.[0]).toMatchObject({ name: 'name', type: 'string' });
    });
  });

  describe('field operations', () => {
    it('should create field with defaults', () => {
      const field = createField();
      expect(field.id).toBeTruthy();
      expect(field.name).toBe('');
      expect(field.type).toBe('string');
      expect(field.children).toEqual([]);
    });

    it('should remove field by id', () => {
      const fields = [
        { id: '1', name: 'a', type: 'string' as const, children: [] },
        { id: '2', name: 'b', type: 'string' as const, children: [] },
      ];
      const result = removeFieldById(fields, '1');
      expect(result).toHaveLength(1);
      expect(result[0].name).toBe('b');
    });

    it('should update field by id', () => {
      const fields = [
        { id: '1', name: 'a', type: 'string' as const, children: [] },
      ];
      const result = updateFieldById(fields, '1', { name: 'updated', type: 'integer' });
      expect(result[0].name).toBe('updated');
      expect(result[0].type).toBe('integer');
    });

    it('should find field by id', () => {
      const fields = [
        {
          id: '1', name: 'parent', type: 'object' as const, children: [
            { id: '2', name: 'child', type: 'string' as const, children: [] },
          ],
        },
      ];
      const found = findFieldById(fields, '2');
      expect(found).toBeTruthy();
      expect(found?.name).toBe('child');
    });
  });
});
```

- [ ] **Step 3: 运行测试验证**

```bash
cd /Users/admin/Desktop/www/dynamic_mcp_go_server/web/admin
npm test -- --testPathPattern=schemaBuilder.test.ts
```

- [ ] **Step 4: 提交代码**

```bash
git add web/admin/src/utils/schemaBuilder.ts web/admin/src/utils/schemaBuilder.test.ts
git commit -m "feat: add schemaBuilder utility for nested schema conversion"
```

---

### Task 2: 创建 SchemaFieldNode.vue 递归组件

**Files:**
- Create: `web/admin/src/components/SchemaFieldNode.vue`

- [ ] **Step 1: 创建 SchemaFieldNode.vue 组件**

```vue
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
```

- [ ] **Step 2: 提交代码**

```bash
git add web/admin/src/components/SchemaFieldNode.vue
git commit -m "feat: add SchemaFieldNode recursive component for nested schema editing"
```

---

### Task 3: 集成到 ServicesPage.vue

**Files:**
- Modify: `web/admin/src/pages/ServicesPage.vue`

- [ ] **Step 1: 添加导入和新状态变量**

在 `<script setup>` 中添加：

```typescript
import SchemaFieldNode from '@/components/SchemaFieldNode.vue';
import {
  fieldsToSchema,
  schemaToFields,
  parseJSONToFields,
  removeFieldById,
  updateFieldById,
  type SchemaField,
} from '@/utils/schemaBuilder';

// Schema editing state - 改为嵌套字段树
const inputFields = ref<SchemaField[]>([]);
const outputFields = ref<SchemaField[]>([]);

// 粘贴解析弹窗状态
const showPasteModal = ref(false);
const pasteTarget = ref<'input' | 'output'>('output');
const pasteJsonStr = ref('');
const pastePreviewFields = ref<SchemaField[]>([]);
const pasteError = ref('');
```

- [ ] **Step 2: 添加字段操作函数**

```typescript
// 添加根字段
const addInputField = () => {
  inputFields.value = [...inputFields.value, createField()];
};
const addOutputField = () => {
  outputFields.value = [...outputFields.value, createField()];
};

// 处理字段更新
const handleInputFieldUpdate = (id: string, updates: Partial<SchemaField>) => {
  inputFields.value = updateFieldById(inputFields.value, id, updates);
};
const handleOutputFieldUpdate = (id: string, updates: Partial<SchemaField>) => {
  outputFields.value = updateFieldById(outputFields.value, id, updates);
};

// 处理字段删除
const handleInputFieldDelete = (id: string) => {
  inputFields.value = removeFieldById(inputFields.value, id);
};
const handleOutputFieldDelete = (id: string) => {
  outputFields.value = removeFieldById(outputFields.value, id);
};

// 粘贴解析相关
const openPasteModal = (target: 'input' | 'output') => {
  pasteTarget.value = target;
  pasteJsonStr.value = '';
  pasteError.value = '';
  pastePreviewFields.value = [];
  showPasteModal.value = true;
};

const parsePasteJson = () => {
  pasteError.value = '';
  pastePreviewFields.value = [];
  try {
    const parsed = JSON.parse(pasteJsonStr.value);
    pastePreviewFields.value = parseJSONToFields(parsed);
  } catch (e) {
    pasteError.value = 'JSON 格式不正确: ' + (e as Error).message;
  }
};

const applyPasteJson = () => {
  if (pasteTarget.value === 'input') {
    inputFields.value = pastePreviewFields.value;
  } else {
    outputFields.value = pastePreviewFields.value;
  }
  showPasteModal.value = false;
};

// 同步函数 - 可视化 → JSON
const syncInputFieldsToSchema = () => {
  inputSchemaStr.value = JSON.stringify(fieldsToSchema(inputFields.value), null, 2);
};
const syncOutputFieldsToSchema = () => {
  outputSchemaStr.value = JSON.stringify(fieldsToSchema(outputFields.value), null, 2);
};

// 同步函数 - JSON → 可视化
const syncInputJsonToFields = () => {
  try {
    const schema = JSON.parse(inputSchemaStr.value);
    inputFields.value = schemaToFields(schema);
    inputSchemaError.value = '';
  } catch {
    inputSchemaError.value = 'JSON 格式不正确';
  }
};
const syncOutputJsonToFields = () => {
  try {
    const schema = JSON.parse(outputSchemaStr.value);
    outputFields.value = schemaToFields(schema);
    outputSchemaError.value = '';
  } catch {
    outputSchemaError.value = 'JSON 格式不正确';
  }
};
```

- [ ] **Step 3: 修改模板中的 Schema 编辑部分**

替换原有的可视化字段编辑区域（约 258-376 行）为：

```vue
<!-- Input Schema (Dual Mode) -->
<div class="space-y-4 pt-4 border-t border-gray-100">
  <h4 class="text-sm font-semibold text-gray-800 flex items-center">
    <svg class="w-4 h-4 mr-1.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/></svg>
    入参 Schema
  </h4>

  <!-- Mode Switch -->
  <div class="flex items-center gap-2">
    <button type="button" @click="inputSchemaMode = 'visual'"
      :class="inputSchemaMode === 'visual' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
      class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors">
      可视化编辑
    </button>
    <button type="button" @click="inputSchemaMode = 'json'; syncInputFieldsToSchema()"
      :class="inputSchemaMode === 'json' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
      class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors">
      JSON 编辑
    </button>
    <button type="button" @click="openPasteModal('input')"
      class="px-3 py-1.5 text-xs font-medium rounded-lg bg-emerald-50 text-emerald-700 hover:bg-emerald-100 transition-colors">
      粘贴 JSON 生成
    </button>
  </div>

  <!-- Visual Mode -->
  <div v-if="inputSchemaMode === 'visual'" class="space-y-2">
    <SchemaFieldNode
      v-for="field in inputFields"
      :key="field.id"
      :field="field"
      :depth="0"
      @update="handleInputFieldUpdate"
      @delete="handleInputFieldDelete"
    />
    <button type="button" @click="addInputField" class="text-sm text-primary-600 hover:text-primary-700 font-medium">
      + 添加字段
    </button>
  </div>

  <!-- JSON Mode -->
  <div v-if="inputSchemaMode === 'json'">
    <textarea v-model="inputSchemaStr" rows="8"
      class="w-full px-3 py-2 border border-gray-300 rounded-lg font-mono text-xs focus:ring-2 focus:ring-primary-500 focus:border-primary-500"></textarea>
    <p v-if="inputSchemaError" class="text-xs text-red-500 mt-1">{{ inputSchemaError }}</p>
    <button type="button" @click="syncInputJsonToFields" class="mt-1 text-xs text-gray-500 hover:text-primary-600">
      ← 从 JSON 同步到可视化
    </button>
  </div>
</div>

<!-- Output Schema (Dual Mode) -->
<div class="space-y-4 pt-4 border-t border-gray-100">
  <h4 class="text-sm font-semibold text-gray-800 flex items-center">
    <svg class="w-4 h-4 mr-1.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/></svg>
    出参 Schema
  </h4>

  <!-- Mode Switch -->
  <div class="flex items-center gap-2">
    <button type="button" @click="outputSchemaMode = 'visual'"
      :class="outputSchemaMode === 'visual' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
      class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors">
      可视化编辑
    </button>
    <button type="button" @click="outputSchemaMode = 'json'; syncOutputFieldsToSchema()"
      :class="outputSchemaMode === 'json' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
      class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors">
      JSON 编辑
    </button>
    <button type="button" @click="openPasteModal('output')"
      class="px-3 py-1.5 text-xs font-medium rounded-lg bg-emerald-50 text-emerald-700 hover:bg-emerald-100 transition-colors">
      粘贴 JSON 生成
    </button>
  </div>

  <!-- Visual Mode -->
  <div v-if="outputSchemaMode === 'visual'" class="space-y-2">
    <SchemaFieldNode
      v-for="field in outputFields"
      :key="field.id"
      :field="field"
      :depth="0"
      @update="handleOutputFieldUpdate"
      @delete="handleOutputFieldDelete"
    />
    <button type="button" @click="addOutputField" class="text-sm text-primary-600 hover:text-primary-700 font-medium">
      + 添加字段
    </button>
  </div>

  <!-- JSON Mode -->
  <div v-if="outputSchemaMode === 'json'">
    <textarea v-model="outputSchemaStr" rows="8"
      class="w-full px-3 py-2 border border-gray-300 rounded-lg font-mono text-xs focus:ring-2 focus:ring-primary-500 focus:border-primary-500"></textarea>
    <p v-if="outputSchemaError" class="text-xs text-red-500 mt-1">{{ outputSchemaError }}</p>
    <button type="button" @click="syncOutputJsonToFields" class="mt-1 text-xs text-gray-500 hover:text-primary-600">
      ← 从 JSON 同步到可视化
    </button>
  </div>
</div>
```

- [ ] **Step 4: 添加粘贴解析弹窗**

在文件末尾（`</template>` 之后）添加：

```vue
<!-- Paste JSON Modal -->
<teleport to="body">
  <transition name="fade">
    <div v-if="showPasteModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex items-center justify-center min-h-screen px-4">
        <div class="fixed inset-0 bg-black bg-opacity-50" @click="showPasteModal = false"></div>
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-3xl max-h-[90vh] overflow-hidden fade-in">
          <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900">粘贴 JSON 生成 Schema</h3>
            <button @click="showPasteModal = false" class="text-gray-400 hover:text-gray-600">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
              </svg>
            </button>
          </div>
          <div class="p-6 overflow-y-auto max-h-[calc(90vh-130px)]">
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">粘贴 JSON 样本</label>
                <textarea v-model="pasteJsonStr" rows="8"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg font-mono text-xs focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  placeholder='{"code": 0, "data": {"error": "", "status_code": 200}}'></textarea>
                <p v-if="pasteError" class="text-xs text-red-500 mt-1">{{ pasteError }}</p>
              </div>
              <div class="flex gap-2">
                <button @click="parsePasteJson" class="px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700">
                  解析
                </button>
              </div>

              <!-- Preview -->
              <div v-if="pastePreviewFields.length > 0">
                <label class="block text-sm font-medium text-gray-700 mb-1">预览生成的字段</label>
                <div class="border border-gray-200 rounded-lg p-4 bg-gray-50 max-h-60 overflow-y-auto">
                  <div v-for="field in pastePreviewFields" :key="field.id">
                    <div class="text-xs font-mono text-gray-700">
                      {{ field.name }}: {{ field.type }}
                      <span v-if="field.children && field.children.length > 0" class="text-gray-400 ml-1">
                        (含 {{ field.children.length }} 个子字段)
                      </span>
                    </div>
                    <div v-if="field.children" class="ml-4 mt-1">
                      <div v-for="child in field.children" :key="child.id" class="text-xs font-mono text-gray-500">
                        └ {{ child.name }}: {{ child.type }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
            <button @click="showPasteModal = false" class="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50">
              取消
            </button>
            <button @click="applyPasteJson" :disabled="pastePreviewFields.length === 0"
              class="px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 disabled:opacity-50">
              应用到 {{ pasteTarget === 'input' ? '入参' : '出参' }} Schema
            </button>
          </div>
        </div>
      </div>
    </div>
  </transition>
</teleport>
```

- [ ] **Step 5: 更新表单提交逻辑**

修改 `handleSubmit` 函数中的 schema 同步部分：

```typescript
const handleSubmit = async () => {
  inputSchemaError.value = ''
  outputSchemaError.value = ''

  // 同步可视化到 JSON
  if (inputSchemaMode.value === 'visual') syncInputFieldsToSchema()
  if (outputSchemaMode.value === 'visual') syncOutputFieldsToSchema()

  const inputSchema = parseJSON(inputSchemaStr.value)
  if (inputSchemaStr.value.trim() && inputSchema === undefined) {
    inputSchemaError.value = '入参 Schema 不是合法的 JSON'
    return
  }
  const outputSchema = parseJSON(outputSchemaStr.value)
  if (outputSchemaStr.value.trim() && outputSchema === undefined) {
    outputSchemaError.value = '出参 Schema 不是合法的 JSON'
    return
  }

  // ...
};
```

- [ ] **Step 6: 更新模态框初始化逻辑**

修改 `openEditModal` 函数中的 schema 解析部分：

```typescript
const openEditModal = (service) => {
  // ...
  // 解析 schema 为嵌套字段树
  const inputSchema = service.input_schema || { type: 'object', properties: {} };
  const outputSchema = service.output_schema || { type: 'object', properties: {} };
  inputSchemaStr.value = JSON.stringify(inputSchema, null, 2);
  outputSchemaStr.value = JSON.stringify(outputSchema, null, 2);
  inputFields.value = schemaToFields(inputSchema);
  outputFields.value = schemaToFields(outputSchema);
  // ...
};

const openCreateModal = () => {
  // ...
  inputFields.value = [];
  outputFields.value = [];
  // ...
};
```

- [ ] **Step 7: 提交代码**

```bash
git add web/admin/src/pages/ServicesPage.vue
git commit -m "feat: integrate nested schema editor into ServicesPage"
```

---

### Task 4: 测试验证

- [ ] **Step 1: 在浏览器中测试创建服务，添加嵌套字段**

访问 ServicesPage，创建或编辑一个服务，测试：
- 添加 object 类型字段，展开后添加子字段
- 添加 array 类型字段，展开后添加子字段
- 删除父字段时子字段一并删除
- 切换到 JSON 模式查看生成的嵌套 JSON Schema
- 粘贴 JSON 样本自动生成嵌套字段

- [ ] **Step 2: 提交最终版本**

```bash
git add -A
git commit -m "feat: complete nested schema editor feature"
```

---

## 自检清单

- [ ] schemaBuilder.ts 中的递归函数正确处理了多层嵌套
- [ ] SchemaFieldNode.vue 支持任意深度的嵌套渲染
- [ ] 粘贴 JSON 解析能正确处理复杂的嵌套结构
- [ ] JSON 模式和可视化模式双向同步正常
- [ ] 字段增删改操作正确更新状态
- [ ] 提交到数据库的 schema 格式正确
- [ ] 前端测试通过
