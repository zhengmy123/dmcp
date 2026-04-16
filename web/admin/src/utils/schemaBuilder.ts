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
  if (!schema) {
    return [];
  }

  // 处理错误的格式：当properties直接包含type和properties字段时
  if (schema.properties && typeof schema.properties === 'object') {
    const props = schema.properties;
    // 检查是否是错误格式（properties中包含type和properties字段）
    if (props.type === 'object' && props.properties) {
      // 跳过错误的type字段，直接处理嵌套的properties
      return schemaToFields(props, parentId);
    }
  }

  if (!schema.properties) {
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
