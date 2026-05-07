import { View } from '@tarojs/components'
import Taro, { useLoad } from '@tarojs/taro'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { Dispatch, SetStateAction } from 'react'
import { useState, useEffect } from 'react'
import { getAddress, createAddress, updateAddress, type AddressPayload } from '@/services/address'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { useAuthStore } from '@/stores/auth'
import { getRegions } from '@/services/region'
import { Button, Cascader, Cell, CellGroup, Form, FormItem, Input, Switch, SafeArea } from '@/ui/nutui'
import { ADDRESS_EDIT_DRAFT_KEY, formatAddress } from '@/utils/address'
import type { Address } from '@/types/biz'
import './index.scss'

interface CascaderNode {
  value: string
  text: string
  leaf: boolean
  children?: CascaderNode[]
}

interface RegionSelection {
  province_code: string
  province: string
  city_code: string
  city: string
  district_code: string
  district: string
  street_code: string
  street: string
}

const emptyRegion: RegionSelection = {
  province_code: '',
  province: '',
  city_code: '',
  city: '',
  district_code: '',
  district: '',
  street_code: '',
  street: '',
}

function toCascaderNode(region: Awaited<ReturnType<typeof getRegions>>[number]): CascaderNode {
  return {
    value: region.code,
    text: region.name,
    leaf: region.has_children === false || region.level >= 4,
  }
}

function applyAddressToForm(
  address: Pick<
    Address,
    | 'name'
    | 'phone'
    | 'detail'
    | 'is_default'
    | 'province_code'
    | 'province'
    | 'city_code'
    | 'city'
    | 'district_code'
    | 'district'
    | 'street_code'
    | 'street'
  >,
  setFormValues: Dispatch<SetStateAction<{ name: string; phone: string; detail: string }>>,
  setIsDefault: Dispatch<SetStateAction<boolean>>,
  setRegionValue: Dispatch<SetStateAction<RegionSelection>>,
  setRegionCodes: Dispatch<SetStateAction<string[]>>,
) {
  setFormValues({
    name: address.name,
    phone: address.phone,
    detail: address.detail,
  })
  setIsDefault(address.is_default)
  setRegionValue({
    province_code: address.province_code ?? '',
    province: address.province ?? '',
    city_code: address.city_code ?? '',
    city: address.city ?? '',
    district_code: address.district_code ?? '',
    district: address.district ?? '',
    street_code: address.street_code ?? '',
    street: address.street ?? '',
  })
  setRegionCodes([
    address.province_code,
    address.city_code,
    address.district_code,
    address.street_code,
  ].filter(Boolean) as string[])
}

export default function AddressEditPage() {
  const isLoggedIn = useAuthStore((state) => state.isLoggedIn)
  const ensureAuth = useAuthGuard()
  const queryClient = useQueryClient()
  const [editId, setEditId] = useState<string | undefined>()
  const isEdit = !!editId
  const [isDefault, setIsDefault] = useState(false)
  const [regionVisible, setRegionVisible] = useState(false)
  const [regionOptions, setRegionOptions] = useState<CascaderNode[]>([])
  const [regionCodes, setRegionCodes] = useState<string[]>([])
  const [regionValue, setRegionValue] = useState<RegionSelection>(emptyRegion)
  const [formValues, setFormValues] = useState({
    name: '',
    phone: '',
    detail: '',
  })

  useLoad((options) => {
    const id = typeof options?.id === 'string' ? options.id : undefined
    setEditId(id)

    if (!id) {
      Taro.removeStorageSync(ADDRESS_EDIT_DRAFT_KEY)
      return
    }

    const draft = Taro.getStorageSync(ADDRESS_EDIT_DRAFT_KEY) as Address | null
    if (draft && draft.id === id) {
      applyAddressToForm(draft, setFormValues, setIsDefault, setRegionValue, setRegionCodes)
    }
  })

  useEffect(() => {
    if (!isLoggedIn) void ensureAuth(undefined, editId ? `/pages/address/edit/index?id=${editId}` : '/pages/address/edit/index')
  }, [editId, ensureAuth, isLoggedIn])

  const { data: rootRegions = [] } = useQuery({
    queryKey: ['regions', 'root'],
    queryFn: () => getRegions(),
  })

  const { data: existing } = useQuery({
    queryKey: ['address', editId],
    queryFn: () => getAddress(editId!),
    enabled: isEdit,
  })

  useEffect(() => {
    setRegionOptions(rootRegions.map(toCascaderNode))
  }, [rootRegions])

  useEffect(() => {
    if (existing) {
      applyAddressToForm(existing, setFormValues, setIsDefault, setRegionValue, setRegionCodes)
      Taro.removeStorageSync(ADDRESS_EDIT_DRAFT_KEY)
    }
  }, [existing])

  const saveMutation = useMutation({
    mutationFn: async () => {
      const payload: AddressPayload = {
        name: formValues.name.trim(),
        phone: formValues.phone.trim(),
        province_code: regionValue.province_code || undefined,
        province: regionValue.province,
        city_code: regionValue.city_code || undefined,
        city: regionValue.city,
        district_code: regionValue.district_code || undefined,
        district: regionValue.district,
        street_code: regionValue.street_code || undefined,
        street: regionValue.street || undefined,
        detail: formValues.detail.trim(),
        is_default: isDefault,
      }
      if (isEdit) {
        await updateAddress(editId!, payload)
      } else {
        await createAddress(payload)
      }
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['addresses'] })
      void Taro.showToast({ title: '保存成功', icon: 'success' })
      void Taro.navigateBack()
    },
    onError: () => {
      void Taro.showToast({ title: '保存失败，请重试', icon: 'none' })
    },
  })

  async function handleSubmit() {
    if (!regionValue.province) {
      void Taro.showToast({ title: '请选择所在地区', icon: 'none' })
      return
    }

    if (!formValues.name.trim()) {
      void Taro.showToast({ title: '请输入收货人姓名', icon: 'none' })
      return
    }

    if (!/^1[3-9]\d{9}$/.test(formValues.phone.trim())) {
      void Taro.showToast({ title: '请输入正确的手机号', icon: 'none' })
      return
    }

    if (!formValues.detail.trim()) {
      void Taro.showToast({ title: '请输入详细地址', icon: 'none' })
      return
    }

    saveMutation.mutate()
  }

  async function handleRegionLoad(node: { value?: string | number }) {
    const regions = await getRegions(String(node.value ?? ''))
    return regions.map(toCascaderNode)
  }

  return (
    <View className='page-shell address-edit-page'>
      <Form className='address-edit-page__form'>
        <CellGroup className='address-edit-page__group'>
          <FormItem label='收货人' name='name'>
            <Input
              placeholder='请输入收货人姓名'
              value={formValues.name}
              onChange={(value) => setFormValues((prev) => ({ ...prev, name: String(value) }))}
            />
          </FormItem>
          <FormItem label='手机号' name='phone'>
            <Input
              type='tel'
              placeholder='请输入手机号码'
              value={formValues.phone}
              onChange={(value) => setFormValues((prev) => ({ ...prev, phone: String(value) }))}
            />
          </FormItem>
          <Cell
            title='所在地区'
            description={regionValue.province
              ? formatAddress(regionValue)
              : '请选择省/市/区/街道'}
            clickable
            onClick={() => {
              if (regionOptions.length === 0) {
                void Taro.showToast({ title: '地区数据加载中', icon: 'none' })
                return
              }
              setRegionVisible(true)
            }}
          />
          <FormItem label='详细地址' name='detail'>
            <Input
              placeholder='楼栋、单元、门牌号等'
              value={formValues.detail}
              onChange={(value) => setFormValues((prev) => ({ ...prev, detail: String(value) }))}
            />
          </FormItem>
          <Cell
            title='设为默认地址'
            extra={<Switch checked={isDefault} onChange={(val) => setIsDefault(val as boolean)} />}
          />
        </CellGroup>
      </Form>

      <Cascader
        visible={regionVisible}
        options={regionOptions}
        value={regionCodes}
        lazy
        onLoad={handleRegionLoad}
        onClose={() => setRegionVisible(false)}
        onChange={(value, pathNodes) => {
          const codes = value.map((item) => String(item))
          const names = pathNodes.map((node) => String(node.text ?? ''))
          setRegionCodes(codes)
          setRegionValue({
            province_code: codes[0] ?? '',
            province: names[0] ?? '',
            city_code: codes[1] ?? '',
            city: names[1] ?? '',
            district_code: codes[2] ?? '',
            district: names[2] ?? '',
            street_code: codes[3] ?? '',
            street: names[3] ?? '',
          })
          setRegionVisible(false)
        }}
      />

      <View className='address-edit-page__footer'>
        <Button
          type='primary'
          block
          loading={saveMutation.isPending}
          onClick={handleSubmit}
        >
          {isEdit ? '保存修改' : '保存地址'}
        </Button>
        <SafeArea position='bottom' />
      </View>
    </View>
  )
}
