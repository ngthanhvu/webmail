export interface Address {
  name: string
  email: string
}

export interface Mailbox {
  name: string
  delimiter: string
  attributes: string[]
  total: number
  unseen: number
}

export interface MessageSummary {
  uid: number
  subject: string
  from: string
  fromEmail: string
  date: string
  size: number
  flags: string[]
  seen: boolean
  flagged: boolean
  snippet: string
}

export interface Attachment {
  section: string
  filename: string
  contentType: string
  size: number
}

export interface Message {
  uid: number
  flags: string[]
  subject: string
  from: Address
  to: Address[]
  cc: Address[]
  date: string
  messageId: string
  inReplyTo: string
  references: string
  text: string
  html: string
  attachments: Attachment[]
  /** chỉ tồn tại phía client sau khi đổi sao từ view chi tiết */
  flagged?: boolean
}

export interface Draft {
  to: string[]
  cc: string[]
  bcc: string[]
  subject: string
  text: string
  inReplyTo: string
  references: string
}

export interface OutgoingAttachment {
  filename: string
  contentType: string
  contentBase64: string
}

export interface LocalAttachment extends OutgoingAttachment {
  size: number
}

export interface SendPayload {
  to: string[]
  cc: string[]
  bcc: string[]
  subject: string
  text: string
  html: string
  inReplyTo: string
  references: string
  attachments: OutgoingAttachment[]
}

export interface FlagsPayload {
  seen?: boolean
  flagged?: boolean
  deleted?: boolean
}

export type BatchAction = 'star' | 'unread' | 'read' | 'delete'

export interface BatchPayload {
  action: BatchAction
  uids: number[]
}

export type ConfirmState = { type: 'message' } | { type: 'batch'; uids: number[] }

export type StarTarget = { uid: number; flagged?: boolean }
