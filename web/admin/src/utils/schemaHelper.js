// 从schema中提取所有字段路径（支持嵌套）
export function extractSchemaFields(schema, prefix = '') {
  const fields = []
  
  if (!schema || !schema.properties) {
    return fields
  }
  
  for (const [name, prop] of Object.entries(schema.properties)) {
    const fullPath = prefix ? `${prefix}.${name}` : name
    fields.push(fullPath)
    
    // 递归处理嵌套对象
    if (prop.type === 'object' && prop.properties) {
      const nestedFields = extractSchemaFields(prop, fullPath)
      fields.push(...nestedFields)
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
