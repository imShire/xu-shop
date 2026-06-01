import { describe, it, expect, vi } from 'vitest'
import { defineComponent, h, nextTick, ref } from 'vue'
import { mount } from '@vue/test-utils'
import ErrorBoundary from '@/components/ErrorBoundary.vue'
import { ClogReporter, setReporter } from '@/utils/clog'

// Element Plus 全局组件 stub
const globalStubs = {
  'el-result': {
    props: ['icon', 'title', 'subTitle'],
    template: '<div class="el-result"><slot name="extra" /></div>',
  },
  'el-button': {
    template: '<button class="el-button" @click="$emit(\'click\')"><slot /></button>',
  },
}

function makeChild(shouldThrow: { value: boolean }) {
  return defineComponent({
    name: 'Boom',
    setup() {
      return () => {
        if (shouldThrow.value) throw new Error('child boom')
        return h('div', { class: 'child-ok' }, 'ok')
      }
    },
  })
}

describe('ErrorBoundary', () => {
  it('renders fallback when child throws and recovers on retry', async () => {
    setReporter(new ClogReporter({ fetchImpl: vi.fn(async () => new Response()) }))
    const shouldThrow = ref(true)
    const Child = makeChild(shouldThrow)
    const wrapper = mount(ErrorBoundary, {
      props: { name: 'unit-test' },
      slots: { default: () => h(Child) },
      global: { stubs: globalStubs },
    })
    await nextTick()
    expect(wrapper.find('.el-result').exists()).toBe(true)
    expect(wrapper.find('.child-ok').exists()).toBe(false)

    shouldThrow.value = false
    await wrapper.find('button').trigger('click')
    await nextTick()
    expect(wrapper.find('.el-result').exists()).toBe(false)
    expect(wrapper.find('.child-ok').exists()).toBe(true)
  })

  it('reports the error to clog', async () => {
    const fetchImpl = vi.fn(async () => new Response())
    const reporter = new ClogReporter({ fetchImpl })
    setReporter(reporter)
    const shouldThrow = ref(true)
    const Child = makeChild(shouldThrow)
    mount(ErrorBoundary, {
      slots: { default: () => h(Child) },
      global: { stubs: globalStubs },
    })
    await nextTick()
    const queued = reporter._peekQueue()
    expect(queued.length).toBe(1)
    expect(queued[0].message).toContain('child boom')
    expect(queued[0].level).toBe('error')
  })
})
