declare module '@wangeditor/editor-for-vue' {
  import { DefineComponent } from 'vue'
  import { IDomEditor, IEditorConfig, IToolbarConfig } from '@wangeditor/editor'

  export const Editor: DefineComponent<{
    modelValue?: string
    defaultConfig?: Partial<IEditorConfig>
    mode?: string
  }>

  export const Toolbar: DefineComponent<{
    editor?: IDomEditor
    defaultConfig?: Partial<IToolbarConfig>
    mode?: string
  }>
}
