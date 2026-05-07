import { useState } from 'react'
import { View, Text, Input } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useMutation } from '@tanstack/react-query'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { useAuthStore } from '@/stores/auth'
import { Avatar, Button, Cell, CellGroup, Popup } from '@/ui/nutui'
import { updateProfile, uploadAvatar } from '@/services/user'
import './index.scss'

export default function SettingsPage() {
  const { isLoggedIn, user, updateUser } = useAuthStore()
  const ensureAuth = useAuthGuard()

  const [nicknamePopup, setNicknamePopup] = useState(false)
  const [nicknameInput, setNicknameInput] = useState('')

  useDidShow(() => {
    if (!isLoggedIn) {
      void ensureAuth(undefined, '/pages/user/settings/index')
    }
  })

  const avatarMutation = useMutation({
    mutationFn: (filePath: string) => uploadAvatar(filePath),
    onSuccess: (updatedUser) => {
      updateUser({ avatar: updatedUser.avatar })
      void Taro.showToast({ title: '头像已更新', icon: 'success' })
    },
    onError: () => {
      void Taro.showToast({ title: '上传失败，请重试', icon: 'none' })
    },
  })

  const nicknameMutation = useMutation({
    mutationFn: (nickname: string) => updateProfile({ nickname }),
    onSuccess: (updatedUser) => {
      updateUser({ nickname: updatedUser.nickname })
      setNicknamePopup(false)
      void Taro.showToast({ title: '昵称已更新', icon: 'success' })
    },
    onError: () => {
      void Taro.showToast({ title: '更新失败，请重试', icon: 'none' })
    },
  })

  const handleChooseAvatar = () => {
    void Taro.chooseImage({
      count: 1,
      sizeType: ['compressed'],
      sourceType: ['album', 'camera'],
      success: ({ tempFilePaths }) => {
        if (tempFilePaths[0]) {
          avatarMutation.mutate(tempFilePaths[0])
        }
      },
    })
  }

  const handleOpenNickname = () => {
    setNicknameInput(user?.nickname ?? '')
    setNicknamePopup(true)
  }

  const handleSaveNickname = () => {
    const trimmed = nicknameInput.trim()
    if (!trimmed) {
      void Taro.showToast({ title: '昵称不能为空', icon: 'none' })
      return
    }
    nicknameMutation.mutate(trimmed)
  }

  if (!isLoggedIn) {
    return <View className='page-shell' />
  }

  const maskedPhone = user?.phone
    ? user.phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
    : '未绑定'

  return (
    <View className='page-shell settings-page'>
      {/* Profile info */}
      <CellGroup title='个人信息' className='settings-page__group'>
        <Cell
          title='头像'
          className='settings-page__avatar-cell'
          extra={
            <Avatar
              size='normal'
              src={user?.avatar ?? undefined}
              className='settings-page__avatar'
            >
              {user?.nickname?.slice(0, 1) ?? '我'}
            </Avatar>
          }
          clickable
          onClick={handleChooseAvatar}
        />
        <Cell
          title='昵称'
          extra={
            <Text className='settings-page__value'>{user?.nickname ?? '未设置'}</Text>
          }
          clickable
          onClick={handleOpenNickname}
        />
        <Cell
          title='手机号'
          extra={
            <Text className='settings-page__value settings-page__value--muted'>{maskedPhone}</Text>
          }
          description={user?.phone ? undefined : '暂不支持换绑'}
        />
      </CellGroup>

      {/* Account security */}
      <CellGroup title='账号安全' className='settings-page__group'>
        <Cell
          title='修改密码'
          extra='>'
          clickable
          onClick={() =>
            void Taro.navigateTo({ url: '/pages/auth/reset-password/index' })
          }
        />
      </CellGroup>

      {/* Addresses */}
      <CellGroup title='收货管理' className='settings-page__group'>
        <Cell
          title='收货地址'
          description='管理常用收货地址'
          extra='>'
          clickable
          onClick={() => void Taro.navigateTo({ url: '/pages/address/list/index' })}
        />
      </CellGroup>

      {/* Nickname edit popup */}
      <Popup
        visible={nicknamePopup}
        position='bottom'
        round
        onClose={() => setNicknamePopup(false)}
      >
        <View className='settings-page__popup'>
          <Text className='settings-page__popup-title'>修改昵称</Text>
          <Input
            className='settings-page__popup-input'
            value={nicknameInput}
            placeholder='请输入新昵称'
            maxlength={20}
            onInput={(e) => setNicknameInput(e.detail.value)}
          />
          <Button
            type='primary'
            className='settings-page__popup-btn'
            loading={nicknameMutation.isPending}
            onClick={handleSaveNickname}
          >
            保存
          </Button>
        </View>
      </Popup>
    </View>
  )
}
