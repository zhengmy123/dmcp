// 从schema中提取所有字段路径（支持嵌套）
export function extractSchemaFields(schema, prefix = '') {
  const fields = []

  if (!schema) {
    return fields
  }

  // 处理 properties
  if (schema.properties && typeof schema.properties === 'object') {
    const props = schema.properties

    // 遍历所有属性
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

// 将schema转换为树形结构
export function schemaToTree(schema, prefix = '') {
  const nodes = []

  if (!schema || !schema.properties) {
    return nodes
  }

  for (const [name, prop] of Object.entries(schema.properties)) {
    const fullPath = prefix ? `${prefix}.${name}` : name
    const node = {
      name,
      path: fullPath,
      type: prop.type || 'unknown',
      description: prop.description || '',
      children: []
    }

    // 递归处理嵌套对象
    if (prop.type === 'object' && prop.properties) {
      node.children = schemaToTree(prop, fullPath)
    }

    nodes.push(node)
  }

  return nodes
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

// 根据路径获取节点的类型
export function getFieldTypeByPath(schema, path) {
  if (!schema || !path) return 'unknown'

  const parts = path.split('.')
  let current = schema

  for (let i = 0; i < parts.length; i++) {
    const part = parts[i]
    if (current.properties && current.properties[part]) {
      current = current.properties[part]
    } else {
      return 'unknown'
    }
  }

  return current.type || 'unknown'
}

// 根据路径获取节点信息
export function getNodeByPath(nodes, path) {
  if (!nodes || !path) return null

  for (const node of nodes) {
    if (node.path === path) return node
    if (node.children && node.children.length > 0) {
      const found = getNodeByPath(node.children, path)
      if (found) return found
    }
  }
  return null
}
