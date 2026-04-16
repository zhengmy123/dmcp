// 从schema中提取所有字段路径（支持嵌套）
export function extractSchemaFields(schema, prefix = '') {
  const fields = []
  
  if (!schema) {
    return fields
  }
  
  // 处理错误的格式：当properties直接包含type和properties字段时
  if (schema.properties && typeof schema.properties === 'object') {
    // 检查是否是错误格式（properties中包含type和properties字段）
    const props = schema.properties
    if (props.type === 'object' && props.properties) {
      // 跳过错误的type字段，直接处理嵌套的properties
      const nestedFields = extractSchemaFields(props, prefix)
      fields.push(...nestedFields)
      return fields
    }
    
    // 正常处理
    for (const [name, prop] of Object.entries(props)) {
      const fullPath = prefix ? `${prefix}.${name}` : name
      fields.push(fullPath)
      
      // 递归处理嵌套对象
      if (prop.type === 'object' && prop.properties) {
        const nestedFields = extractSchemaFields(prop, fullPath)
        fields.push(...nestedFields)
      }
    }
  }
  
  return fields
}

// 从点号路径创建嵌套对象
export function createNestedObject(path, value) {
  const parts = path.split('.')
  const result = {}
  let current = result
  
  for (let i = 0; i < parts.length; i++) {
    const part = parts[i]
    if (i === parts.length - 1) {
      current[part] = value
    } else {
      current[part] = {}
      current = current[part]
    }
  }
  
  return result
}

// 合并嵌套对象
export function mergeNestedObjects(target, source) {
  for (const key in source) {
    if (typeof source[key] === 'object' && source[key] !== null) {
      if (!target[key]) {
        target[key] = {}
      }
      mergeNestedObjects(target[key], source[key])
    } else {
      target[key] = source[key]
    }
  }
  return target
}
