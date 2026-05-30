import { View, Text, Image } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useState, useRef } from 'react'
import { uploadAftersaleEvidence } from '@/services/aftersale'
import { showErrorToast } from '@/utils/error'
import { isWeapp } from '@/utils/platform'
import './EvidenceUploader.scss'

interface Props {
  value: string[]
  onChange: (next: string[]) => void
  max?: number
  disabled?: boolean
}

const ALLOWED_MIME = ['image/jpeg', 'image/png', 'image/webp']
const ALLOWED_EXT = /\.(jpe?g|png|webp)$/i

export default function EvidenceUploader({
  value,
  onChange,
  max = 6,
  disabled = false,
}: Props) {
  const [uploading, setUploading] = useState(false)
  const inputRef = useRef<HTMLInputElement | null>(null)

  async function uploadOne(filePath: string) {
    setUploading(true)
    try {
      const url = await uploadAftersaleEvidence(filePath)
      onChange([...value, url])
    } catch (err) {
      showErrorToast(err, '上传失败，请稍后再试')
    } finally {
      setUploading(false)
    }
  }

  function handleChoose() {
    if (disabled || uploading) return
    if (value.length >= max) {
      void Taro.showToast({ title: `最多 ${max} 张`, icon: 'none' })
      return
    }

    if (isWeapp) {
      void Taro.chooseImage({
        count: Math.min(max - value.length, 9),
        sizeType: ['compressed'],
        sourceType: ['album', 'camera'],
        success: async (res) => {
          for (const path of res.tempFilePaths) {
            await uploadOne(path)
          }
        },
      })
      return
    }
    inputRef.current?.click()
  }

  async function handleH5Files(files: FileList | null) {
    if (!files) return
    for (let i = 0; i < files.length && value.length + i < max; i++) {
      const f = files[i]
      if (!ALLOWED_MIME.includes(f.type) || !ALLOWED_EXT.test(f.name)) {
        void Taro.showToast({ title: '仅支持 jpg/png/webp', icon: 'none' })
        continue
      }
      const url = URL.createObjectURL(f)
      await uploadOne(url)
    }
    if (inputRef.current) inputRef.current.value = ''
  }

  function handleRemove(idx: number) {
    const next = value.slice()
    next.splice(idx, 1)
    onChange(next)
  }

  return (
    <View className='evidence-uploader'>
      {value.map((url, idx) => (
        <View className='evidence-uploader__tile' key={`${url}-${idx}`}>
          <Image className='evidence-uploader__img' src={url} mode='aspectFill' />
          {!disabled && (
            <Text
              className='evidence-uploader__remove'
              onClick={() => handleRemove(idx)}
            >
              ×
            </Text>
          )}
        </View>
      ))}
      {!disabled && value.length < max && (
        <View
          className='evidence-uploader__tile evidence-uploader__add'
          onClick={handleChoose}
        >
          <Text className='evidence-uploader__add-icon'>+</Text>
          <Text className='evidence-uploader__add-text'>
            {uploading ? '上传中' : `凭证 ${value.length}/${max}`}
          </Text>
          {!isWeapp && (
            <input
              ref={inputRef}
              type='file'
              accept='image/jpeg,image/png,image/webp'
              multiple
              style={{ display: 'none' }}
              onChange={(e) => void handleH5Files(e.target.files)}
            />
          )}
        </View>
      )}
    </View>
  )
}
