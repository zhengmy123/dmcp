<template>
  <div class="schema-field-tree-node">
    <div
      class="flex items-center py-1 px-2 cursor-pointer hover:bg-gray-100 rounded transition-colors"
      :class="{ 'bg-primary-50': isSelected }"
      @click="handleClick"
    >
      <!-- 展开/折叠图标 -->
      <span
        v-if="hasChildren"
        class="w-4 h-4 flex items-center justify-center mr-1 text-gray-400 hover:text-gray-600"
        @click.stop="toggleExpand"
      >
        <svg
          class="w-3 h-3 transform transition-transform"
          :class="{ 'rotate-90': isExpanded }"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </span>
      <span v-else class="w-4 h-4 mr-1"></span>

      <!-- 字段图标 -->
      <span class="mr-2 text-gray-400">
        <svg v-if="node.type === 'object'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
        </svg>
        <svg v-else-if="node.type === 'array'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
        </svg>
        <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
        </svg>
      </span>

      <!-- 字段名 -->
      <span
        class="text-sm"
        :class="isSelected ? 'text-primary-700 font-medium' : 'text-gray-700'"
      >
        {{ node.name }}
      </span>

      <!-- 类型标签 -->
      <span class="ml-2 text-xs text-gray-400">{{ node.type }}</span>
    </div>

    <!-- 子节点 -->
    <div
      v-if="hasChildren && isExpanded"
      class="ml-4 border-l border-gray-200 pl-2"
    >
      <SchemaFieldTreeNode
        v-for="child in node.children"
        :key="child.path"
        :node="child"
        :selected-path="selectedPath"
        @select="handleSelect"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  node: {
    type: Object,
    required: true
  },
  selectedPath: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['select'])

const isExpanded = ref(true)

const hasChildren = computed(() => {
  return props.node.children && props.node.children.length > 0
})

const isSelected = computed(() => {
  return props.selectedPath === props.node.path
})

const toggleExpand = () => {
  isExpanded.value = !isExpanded.value
}

const handleClick = () => {
  emit('select', props.node.path)
}

const handleSelect = (path) => {
  emit('select', path)
}
</script>
