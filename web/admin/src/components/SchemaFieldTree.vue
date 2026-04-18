<template>
  <div class="schema-field-tree border border-gray-300 rounded-lg p-2 max-h-48 overflow-y-auto bg-white">
    <div v-if="nodes.length === 0" class="text-sm text-gray-500 py-2 px-3">
      暂无可用字段
    </div>
    <template v-else>
      <SchemaFieldTreeNode
        v-for="node in nodes"
        :key="node.path"
        :node="node"
        :selected-path="selectedPath"
        @select="handleSelect"
      />
    </template>
  </div>
</template>

<script setup>
import SchemaFieldTreeNode from './SchemaFieldTreeNode.vue'

const props = defineProps({
  nodes: {
    type: Array,
    default: () => []
  },
  selectedPath: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['select'])

const handleSelect = (path) => {
  emit('select', path)
}
</script>
