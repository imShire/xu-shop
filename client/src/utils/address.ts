interface AddressParts {
  province?: string | null
  city?: string | null
  district?: string | null
  street?: string | null
  detail?: string | null
}

export const ADDRESS_EDIT_DRAFT_KEY = 'address-edit-draft'

export function formatAddress(parts: AddressParts) {
  return [
    parts.province,
    parts.city,
    parts.district,
    parts.street,
    parts.detail,
  ].filter(Boolean).join('')
}
