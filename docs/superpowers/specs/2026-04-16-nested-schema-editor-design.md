# 多层嵌套 Schema 编辑器设计

## 1. 背景与目标

当前 HTTP 服务管理界面的入参/出参 Schema 编辑器仅支持单层扁平结构，无法处理多层嵌套的 JSON 响应。

**目标**：
1. 支持多层嵌套 Schema 的可视化编辑
2. 支持粘贴样本 JSON 自动构建 Schema
3. 保持与现有 JSON 编辑模式的双向同步

---

## 2. 数据结构

### 2.1 嵌套字段树节点

```typescript
interface SchemaField {
  id: string;           // 唯一标识，用于 Vue key
  name: string;         // 字段名
  type: 'string' | 'integer' | 'number' | 'boolean' | 'array' | 'object';
  description?: string;
  required?: boolean;
  default?: any;        // 默认值
  children?: SchemaField[];  // 子字段（仅 object/array 类型有）
  expanded?: boolean;  // UI 折叠状态
}
```

### 2.2 SchemaField 与 JSON Schema 的映射关系

| SchemaField.type | JSON Schema |
|------------------|-------------|
| `string` | `{ "type": "string" }` |
| `integer` | `{ "type": "integer" }` |
| `number` | `{ "type": "number" }` |
| `boolean` | `{ "type": "boolean" }` |
| `object` | `{ "type": "object", "properties": {...} }` |
| `array` | `{ "type": "array", "items": { "type": "object", "properties": {...} } }` |

---

## 3. 核心算法

### 3.1 `fieldsToSchema` - 嵌套字段树 → JSON Schema（递归）

```typescript
function fieldsToSchema(fields: SchemaField[]): any {
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
```

### 3.2 `schemaToFields` - JSON Schema → 嵌套字段树（递归）

```typescript
function schemaToFields(schema: any, parentId = ''): SchemaField[] {
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
      default: formatDefault(propObj.default),
      children,
      expanded: true,
    });
  }

  return fields;
}
```

### 3.3 `parseJSONToFields` - 样本 JSON → 嵌套字段树（递归）

```typescript
function parseJSONToFields(obj: any, parentId = ''): SchemaField[] {
  if (obj === null || obj === undefined) {
    return [];
  }

  if (Array.isArray(obj)) {
    // 取第一个元素推断结构
    if (obj.length > 0) {
      return [{
        id: parentId || 'root',
        name: 'items',
        type: 'array',
        children: parseJSONToFields(obj[0], parentId ? `${parentId}-items` : 'items'),
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
    let type = inferType(value);
    let children: SchemaField[] | undefined;

    if (type === 'object' && typeof value === 'object') {
      children = parseJSONToFields(value, id);
    } else if (type === 'array') {
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

function inferType(value: any): SchemaField['type'] {
  if (typeof value === 'string') return 'string';
  if (typeof value === 'number') {
    return Number.isInteger(value) ? 'integer' : 'number';
  }
  if (typeof value === 'boolean') return 'boolean';
  if (Array.isArray(value)) return 'array';
  if (typeof value === 'object' && value !== null) return 'object';
  return 'string';
}
```

---

## 4. 前端组件设计

### 4.1 组件结构

```
SchemaEditor (主组件)
├── ModeSwitch (模式切换：可视化 / JSON)
├── VisualEditor (可视化编辑模式)
│   ├── FieldTree (字段树)
│   │   └── FieldNode (递归字段节点)
│   │       ├── FieldRow (字段行)
│   │       ├── ChildrenContainer (子字段容器，递归)
│   │       └── AddChildButton (添加子字段)
│   └── AddRootFieldButton
├── JsonEditor (JSON 编辑模式)
│   └── textarea + sync button
└── PasteParser (粘贴解析模式)
    └── textarea + parse button + preview
```

### 4.2 FieldNode 递归组件

```vue
<template>
  <div class="field-node" :class="{ 'is-nested': depth > 0 }">
    <div class="field-row">
      <!-- 缩进指示 -->
      <span v-if="depth > 0" class="indent-guide" :style="{ left: `${(depth - 1) * 16 + 8}px` }"></span>

      <!-- 展开/折叠按钮 -->
      <button
        v-if="field.type === 'object' || field.type === 'array'"
        @click="field.expanded = !field.expanded"
        class="expand-btn"
      >
        {{ field.expanded ? '▼' : '▶' }}
      </button>
      <span v-else class="expand-placeholder"></span>

      <!-- 字段名 -->
      <input v-model="field.name" class="field-name" placeholder="字段名" />

      <!-- 类型选择 -->
      <select v-model="field.type" class="field-type" @change="onTypeChange">
        <option value="string">字符串</option>
        <option value="integer">整数</option>
        <option value="number">数字</option>
        <option value="boolean">布尔</option>
        <option value="array">数组</option>
        <option value="object">对象</option>
      </select>

      <!-- 描述 -->
      <input v-model="field.description" class="field-desc" placeholder="描述" />

      <!-- 必填 -->
      <label class="field-required">
        <input v-model="field.required" type="checkbox" />
        必填
      </label>

      <!-- 默认值 -->
      <input
        v-model="field.default"
        class="field-default"
        placeholder="默认值"
        :disabled="field.type === 'object' || field.type === 'array'"
      />

      <!-- 操作按钮 -->
      <button @click="$emit('delete', field.id)" class="delete-btn" title="删除">✕</button>
    </div>

    <!-- 子字段 -->
    <div v-if="(field.type === 'object' || field.type === 'array') && field.expanded" class="children">
      <FieldNode
        v-for="child in field.children"
        :key="child.id"
        :field="child"
        :depth="depth + 1"
        @update="handleChildUpdate"
        @delete="handleChildDelete"
      />
      <button @click="addChild" class="add-child-btn">+ 添加子字段</button>
    </div>
  </div>
</template>
```

---

## 5. UI 交互流程

### 5.1 可视化编辑模式

1. **添加根字段**：点击「+ 添加字段」在根层级添加新字段
2. **展开/折叠**：点击 `▶`/`▼` 按钮展开或折叠 object/array 的子字段
3. **添加子字段**：展开后显示「+ 添加子字段」按钮，点击后在当前父节点下添加子字段
4. **修改类型**：选择 `object` 或 `array` 时自动显示子字段区域
5. **删除字段**：点击 `✕` 删除字段（同时删除所有子字段）

### 5.2 粘贴解析模式

1. 用户粘贴一段 JSON 示例到文本框
2. 点击「解析」按钮
3. 系统调用 `parseJSONToFields` 生成嵌套字段树
4. 预览生成的字段结构（可编辑）
5. 确认后合并到可视化编辑器

### 5.3 JSON 编辑模式

1. 切换到 JSON 模式后显示原始 JSON Schema
2. 可直接编辑 JSON 文本
3. 点击「← 同步到可视化」将 JSON 解析为字段树
4. 编辑后自动同步回 JSON 文本

---

## 6. 模式切换逻辑

```
┌──────────────────────────────────────────────────────┐
│  当前模式                                              │
├──────────────────────────────────────────────────────┤
│  可视化 ←→ JSON (双向同步)                            │
│    ↓                                                  │
│  粘贴解析 → 可视化 (单向，解析后进入可视化编辑)         │
└──────────────────────────────────────────────────────┘
```

| 切换操作 | 数据流向 | 说明 |
|---------|---------|------|
| 可视化 → JSON | `fieldsToSchema` | 序列化当前字段树为 JSON Schema |
| JSON → 可视化 | `schemaToFields` | 解析 JSON Schema 为字段树 |
| 粘贴解析 → 可视化 | `parseJSONToFields` | 从样本 JSON 生成字段树 |

---

## 7. 与现有代码的集成

### 7.1 ServicesPage.vue 改动

- 保留 `inputFields` / `outputFields` 变量，类型从 `Array<SimpleField>` 改为 `Array<SchemaField>`
- `fieldsToSchema` 函数替换为递归版本
- `schemaToFields` 函数替换为递归版本
- 新增 `parseJSONToFields` 函数
- 组件模板中 `FieldNode` 替换原有的 flat field rendering
- 新增「粘贴解析」按钮和弹窗

### 7.2 新增文件

| 文件 | 说明 |
|------|------|
| `web/admin/src/components/SchemaFieldNode.vue` | 递归字段节点组件 |
| `web/admin/src/utils/schemaBuilder.ts` | `fieldsToSchema`、`schemaToFields`、`parseJSONToFields` 函数 |

---

## 8. Schema 双向同步示例

### 8.1 JSON Schema → 嵌套字段树

**输入**:
```json
{
  "type": "object",
  "properties": {
    "code": { "type": "integer" },
    "data": {
      "type": "object",
      "properties": {
        "error": { "type": "string" },
        "status_code": { "type": "integer" },
        "stdout": { "type": "string" }
      }
    },
    "message": { "type": "string" }
  },
  "required": ["code", "message"]
}
```

**输出**:
```
root
├── code (integer, required)
├── data (object)
│   ├── error (string)
│   ├── status_code (integer)
│   └── stdout (string)
└── message (string, required)
```

### 8.2 嵌套字段树 → JSON Schema

**输入**:
```
root
├── code (integer, required)
├── data (object)
│   ├── error (string)
│   └── stdout (string)
└── message (string, required)
```

**输出**:
```json
{
  "type": "object",
  "properties": {
    "code": { "type": "integer" },
    "data": {
      "type": "object",
      "properties": {
        "error": { "type": "string" },
        "stdout": { "type": "string" }
      }
    },
    "message": { "type": "string" }
  },
  "required": ["code", "message"]
}
```

### 8.3 样本 JSON → 嵌套字段树

**输入**:
```json
{
  "code": 0,
  "data": {
    "error": "",
    "status_code": 200,
    "stdout": "123"
  },
  "message": "success"
}
```

**输出**:
```
root
├── code (integer)
├── data (object)
│   ├── error (string)
│   ├── status_code (integer)
│   └── stdout (string)
└── message (string)
```

---

## 9. 特殊处理

### 9.1 空 Schema 默认值

- `schemaToFields({ type: 'object', properties: {} })` → `[]`（空数组）
- `schemaToFields(null)` → `[]`
- `fieldsToSchema([])` → `{ type: 'object', properties: {} }`

### 9.2 类型不一致

- `type: 'array'` 的 `items` 如果是 `object` 类型，继续递归解析
- `type: 'array'` 的 `items` 如果是简单类型（`string`/`integer` 等），不生成子字段

### 9.3 字段名冲突

- 同一父节点下不允许同名子字段
- 根层级字段名不能为空

---

## 10. 实现顺序

1. **工具函数**：`schemaBuilder.ts` 中实现 `fieldsToSchema`、`schemaToFields`、`parseJSONToFields`
2. **递归组件**：`SchemaFieldNode.vue` 实现 FieldNode 递归组件
3. **集成**：在 `ServicesPage.vue` 中替换原有字段编辑逻辑
4. **粘贴解析弹窗**：新增粘贴解析交互

---

## 11. 测试验证

### 11.1 单元测试

- `fieldsToSchema`：嵌套字段树 → JSON Schema 正确
- `schemaToFields`：JSON Schema → 嵌套字段树正确
- `parseJSONToFields`：样本 JSON → 嵌套字段树正确
- 双向转换保持数据一致（`schemaToFields(fieldsToSchema(x)) === x`）

### 11.2 集成测试

- 在 ServicesPage 中创建多层嵌套服务，保存后刷新，数据保持一致
- 粘贴解析后编辑，保存后刷新，数据保持一致
