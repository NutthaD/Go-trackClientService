import { useEffect, useMemo, useState } from 'react'
import type { FormEvent, ReactNode } from 'react'
import { getClients, updateService } from './api'
import type { Client, Service } from './types'
import './App.css'

type ServiceForm = {
  serv_name: string
  uat_status: string
  tested_by: string
  pos_os_nss: string
  testcase_result_link: string
  week: string
  uat_submitted_date: string
  service_type: string
  production_status: string
}

type ServiceRow = {
  client: Client
  service: Service
}

type SortKey = 'client_id' | 'client_short_name' | 'service_id' | 'serv_name'
type SortDirection = 'asc' | 'desc'

const UAT_STATUSES = [
  'Waiting test',
  'Working',
  'Block',
  'Block by defect',
  'Defect',
  'Pass',
  'Cancel',
  'Hold testing',
] as const

const PRODUCTION_STATUSES = [
  { value: 'NOT_PROD', label: 'ยังไม่ขึ้น Production' },
  { value: 'PROD', label: 'ขึ้น Production แล้ว' },
  { value: 'NOT_APPLICABLE', label: 'ไม่เกี่ยวข้อง' },
] as const

function emptyForm(service: Service): ServiceForm {
  return {
    serv_name: service.serv_name,
    uat_status: service.uat_status,
    tested_by: service.tested_by ?? '',
    pos_os_nss: service.pos_os_nss ?? '',
    testcase_result_link: service.testcase_result_link ?? '',
    week: service.week ?? '',
    uat_submitted_date: formatDateForInput(service.uat_submitted_date),
    service_type: service.service_type ?? '',
    production_status: service.production_status,
  }
}

function formatDateForInput(value: string | null) {
  if (!value) return ''
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value)
  return match ? `${match[3]}/${match[2]}/${match[1]}` : value
}

function formatDateForAPI(value: string) {
  const trimmedValue = value.trim()
  if (!trimmedValue) return ''

  const match = /^(\d{2})\/(\d{2})\/(\d{4})$/.exec(trimmedValue)
  if (!match) return null

  const [, day, month, year] = match
  const date = new Date(Date.UTC(Number(year), Number(month) - 1, Number(day)))
  if (
    date.getUTCFullYear() !== Number(year) ||
    date.getUTCMonth() !== Number(month) - 1 ||
    date.getUTCDate() !== Number(day)
  ) {
    return null
  }
  return `${year}-${month}-${day}`
}

function statusClass(status: string) {
  switch (status) {
    case 'Pass':
      return 'status-pill status-pill--green'
    case 'Block':
    case 'Defect':
      return 'status-pill status-pill--pink'
    case 'Block by defect':
      return 'status-pill status-pill--yellow'
    case 'Working':
    case 'Hold testing':
      return 'status-pill status-pill--blue'
    case 'Cancel':
      return 'status-pill status-pill--gray'
    default:
      return 'status-pill status-pill--purple'
  }
}

function App() {
  const [clients, setClients] = useState<Client[]>([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('ทั้งหมด')
  const [sort, setSort] = useState<{ key: SortKey; direction: SortDirection }>({
    key: 'client_id',
    direction: 'asc',
  })
  const [selectedRow, setSelectedRow] = useState<ServiceRow | null>(null)
  const [form, setForm] = useState<ServiceForm | null>(null)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState('')
  const [notice, setNotice] = useState('')

  async function loadClients() {
    setLoading(true)
    setLoadError('')
    try {
      setClients(await getClients())
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : 'โหลดข้อมูลไม่สำเร็จ')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    getClients()
      .then(setClients)
      .catch((error: unknown) => {
        setLoadError(error instanceof Error ? error.message : 'โหลดข้อมูลไม่สำเร็จ')
      })
      .finally(() => setLoading(false))
  }, [])

  const allRows = useMemo(
    () =>
      clients.flatMap((client) =>
        client.services.map((service) => ({ client, service })),
      ),
    [clients],
  )

  const visibleRows = useMemo(() => {
    const query = search.trim().toLocaleLowerCase()
    return allRows.filter(({ client, service }) => {
      const matchesStatus =
        statusFilter === 'ทั้งหมด' || service.uat_status === statusFilter
      const searchableText = [
        client.client_id,
        client.client_short_name,
        service.client_service_id,
        service.serv_name,
        service.tested_by ?? '',
        service.service_type ?? '',
      ]
        .join(' ')
        .toLocaleLowerCase()
      return matchesStatus && (!query || searchableText.includes(query))
    })
  }, [allRows, search, statusFilter])

  const sortedRows = useMemo(() => {
    const isClientSort = sort.key === 'client_id' || sort.key === 'client_short_name'
    const clientKey: 'client_id' | 'client_short_name' =
      sort.key === 'client_id' || sort.key === 'client_short_name'
        ? sort.key
        : 'client_id'
    const serviceKey: 'service_id' | 'serv_name' = sort.key === 'service_id' || sort.key === 'serv_name'
      ? sort.key
      : 'service_id'

    return [...visibleRows].sort((a, b) => {
      const clientOrder = compareText(
        a.client[clientKey],
        b.client[clientKey],
        isClientSort ? sort.direction : 'asc',
      )
      if (clientOrder !== 0) return clientOrder

      const clientIDOrder = compareText(a.client.client_id, b.client.client_id, 'asc')
      if (clientIDOrder !== 0) return clientIDOrder

      return compareText(
        a.service[serviceKey],
        b.service[serviceKey],
        !isClientSort ? sort.direction : 'asc',
      )
    })
  }, [visibleRows, sort])

  const readyForProductionCount = allRows.filter(
    ({ service }) =>
      service.uat_status === 'Pass' && service.production_status === 'NOT_PROD',
  ).length
  const waitingCount = allRows.filter(
    ({ service }) => service.uat_status === 'Waiting test',
  ).length
  const workingCount = allRows.filter(
    ({ service }) => service.uat_status === 'Working',
  ).length
  const blockCount = allRows.filter(
    ({ service }) => service.uat_status === 'Block',
  ).length
  const blockByDefectCount = allRows.filter(
    ({ service }) => service.uat_status === 'Block by defect',
  ).length
  const defectCount = allRows.filter(
    ({ service }) => service.uat_status === 'Defect',
  ).length
  const productionCount = allRows.filter(
    ({ service }) => service.production_status === 'PROD',
  ).length

  function openEditor(row: ServiceRow) {
    setSelectedRow(row)
    setForm(emptyForm(row.service))
    setSaveError('')
  }

  function closeEditor() {
    if (!saving) {
      setSelectedRow(null)
      setForm(null)
      setSaveError('')
    }
  }

  async function saveService(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!selectedRow || !form) return

    const submittedDate = formatDateForAPI(form.uat_submitted_date)
    if (submittedDate === null) {
      setSaveError('วันที่ส่ง UAT ต้องเป็นรูปแบบ dd/mm/yyyy และเป็นวันที่ที่ถูกต้อง')
      return
    }

    setSaving(true)
    setSaveError('')
    try {
      await updateService(
        selectedRow.client.client_id,
        selectedRow.service.client_service_id,
        { ...form, uat_submitted_date: submittedDate },
      )
      setNotice(`บันทึกข้อมูล ${selectedRow.service.client_service_id} แล้ว`)
      setSelectedRow(null)
      setForm(null)
      await loadClients()
      window.setTimeout(() => setNotice(''), 3500)
    } catch (error) {
      setSaveError(error instanceof Error ? error.message : 'บันทึกข้อมูลไม่สำเร็จ')
    } finally {
      setSaving(false)
    }
  }

  function updateFormField(field: keyof ServiceForm, value: string) {
    setForm((current) => (current ? { ...current, [field]: value } : current))
  }

  function toggleSort(key: SortKey) {
    setSort((current) => ({
      key,
      direction: current.key === key && current.direction === 'asc' ? 'desc' : 'asc',
    }))
  }

  return (
    <main className="app-shell">
      <header className="topbar">
        <a className="brand" href="/" aria-label="Track Client Service หน้าหลัก">
          <span className="brand-mark" aria-hidden="true">
            <svg viewBox="0 0 32 32" fill="none">
              <path
                d="M7 22.5 13 16l4.2 4.2L25 11"
                stroke="currentColor"
                strokeWidth="2.5"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
              <path
                d="M19.5 11H25v5.5"
                stroke="currentColor"
                strokeWidth="2.5"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
          </span>
          <span className="brand-name">track<span>client</span></span>
        </a>
        <div className="topbar-right">
          <span className={`connection-dot ${loadError ? 'connection-dot--offline' : loading ? 'connection-dot--loading' : ''}`} />
          <span>{loadError ? 'API ยังไม่เชื่อมต่อ' : loading ? 'กำลังเชื่อมต่อ API' : 'เชื่อมต่อ API แล้ว'}</span>
          <span className="topbar-avatar" aria-hidden="true">T</span>
        </div>
      </header>

      <section className="page-content">
        <div className="page-heading">
          <div>
            <p className="eyebrow">CLIENT SERVICE WORKSPACE</p>
            <h1>ภาพรวมการทดสอบ</h1>
            <p className="page-subtitle">
              ติดตามสถานะ UAT และข้อมูล Service ของผู้ว่าจ้าง
            </p>
          </div>
          <button className="button button--soft" type="button" onClick={() => void loadClients()} disabled={loading}>
            <RefreshIcon />
            {loading ? 'กำลังโหลด...' : 'รีเฟรชข้อมูล'}
          </button>
        </div>

        <section className="stats-grid" aria-label="สรุปสถานะ">
          <StatCard label="บริการทั้งหมด" value={allRows.length} detail="Service List" tone="purple" icon={<GridIcon />} />
          <ProgressStatCard waitingCount={waitingCount} workingCount={workingCount} />
          <IssueStatCard
            blockCount={blockCount}
            blockByDefectCount={blockByDefectCount}
            defectCount={defectCount}
          />
          <ReleaseStatCard
            readyCount={readyForProductionCount}
            productionCount={productionCount}
          />
        </section>

        <section className="table-panel">
          <div className="panel-heading">
            <div>
              <h2>Service List</h2>
              <p>กำลังแสดง {sortedRows.length} จาก {allRows.length} รายการ · จัดกลุ่มตามผู้ว่าจ้าง</p>
            </div>
          </div>

          <div className="table-toolbar">
            <label className="search-field">
              <SearchIcon />
              <input
                aria-label="ค้นหาผู้ว่าจ้างหรือบริการ"
                placeholder="ค้นหาชื่อผู้ว่าจ้าง, service, ผู้ทดสอบ..."
                value={search}
                onChange={(event) => setSearch(event.target.value)}
              />
              {search && (
                <button className="clear-search" type="button" onClick={() => setSearch('')} aria-label="ล้างคำค้นหา">
                  ×
                </button>
              )}
            </label>
            <label className="filter-field">
              <span>สถานะ</span>
              <select value={statusFilter} onChange={(event) => setStatusFilter(event.target.value)}>
                <option>ทั้งหมด</option>
                {UAT_STATUSES.map((status) => <option key={status}>{status}</option>)}
              </select>
            </label>
          </div>

          {loadError ? (
            <div className="state-panel state-panel--error" role="alert">
              <span className="state-icon"><AlertIcon /></span>
              <h3>ยังโหลดข้อมูลไม่ได้</h3>
              <p>{loadError}</p>
              <p className="state-help">ตรวจสอบว่า Go API และ MySQL เปิดอยู่ แล้วลองอีกครั้ง</p>
              <button className="button button--primary" type="button" onClick={() => void loadClients()}>
                ลองโหลดอีกครั้ง
              </button>
            </div>
          ) : loading ? (
            <div className="state-panel">
              <span className="loading-spinner" />
              <p>กำลังโหลด Service List...</p>
            </div>
          ) : visibleRows.length === 0 ? (
            <div className="state-panel">
              <span className="state-icon state-icon--purple"><SearchIcon /></span>
              <h3>{allRows.length === 0 ? 'ยังไม่มีข้อมูลบริการ' : 'ไม่พบรายการที่ค้นหา'}</h3>
              <p>{allRows.length === 0 ? 'เมื่อมีข้อมูลผู้ว่าจ้างและบริการแล้ว รายการจะแสดงที่นี่' : 'ลองเปลี่ยนคำค้นหาหรือตัวกรองสถานะ'}</p>
            </div>
          ) : (
            <div className="table-scroll">
              <table>
                <thead>
                  <tr>
                    <th><SortableHeader label="Client ID" sortKey="client_id" sort={sort} onSort={toggleSort} /></th>
                    <th><SortableHeader label="ผู้ว่าจ้าง" sortKey="client_short_name" sort={sort} onSort={toggleSort} /></th>
                    <th><SortableHeader label="Service ID" sortKey="service_id" sort={sort} onSort={toggleSort} /></th>
                    <th><SortableHeader label="Service" sortKey="serv_name" sort={sort} onSort={toggleSort} /></th>
                    <th>สถานะ UAT</th>
                    <th>ผู้ทดสอบ</th>
                    <th>Week</th>
                    <th>วันที่ส่ง UAT</th>
                    <th>Production</th>
                    <th><span className="sr-only">จัดการ</span></th>
                  </tr>
                </thead>
                <tbody>
                  {sortedRows.map(({ client, service }, index) => {
                    const isClientStart = index === 0 || sortedRows[index - 1].client.client_id !== client.client_id
                    return (
                    <tr className={isClientStart ? 'client-group-start' : ''} key={service.client_service_id}>
                      <td className="client-id-cell">{isClientStart ? client.client_id : ''}</td>
                      <td className="client-name-cell">{isClientStart ? client.client_short_name : ''}</td>
                      <td className="service-id-cell">{service.service_id}</td>
                      <td className="service-name-cell">{service.serv_name}</td>
                      <td><span className={statusClass(service.uat_status)}>{service.uat_status}</span></td>
                      <td>{service.tested_by || <span className="muted-value">ยังไม่ระบุ</span>}</td>
                      <td>{service.week || <span className="muted-value">—</span>}</td>
                      <td>{formatDate(service.uat_submitted_date)}</td>
                      <td><ProductionBadge status={service.production_status} /></td>
                      <td className="action-cell">
                        <button className="edit-button" type="button" onClick={() => openEditor({ client, service })}>
                          แก้ไข
                        </button>
                      </td>
                    </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          )}
        </section>

        <footer className="page-footer">
          <span>Track Client Service</span>
          <span>ข้อมูลจะแสดงตามรายการที่บันทึกในระบบ</span>
        </footer>
      </section>

      {notice && <div className="toast" role="status"><CheckIcon />{notice}</div>}

      {selectedRow && form && (
        <div className="modal-backdrop" role="presentation" onMouseDown={(event) => {
          if (event.target === event.currentTarget) closeEditor()
        }}>
          <section className="edit-modal" role="dialog" aria-modal="true" aria-labelledby="edit-title">
            <div className="modal-heading">
              <div>
                <p className="eyebrow">แก้ไขข้อมูลบริการ</p>
                <h2 id="edit-title">{selectedRow.service.client_service_id}</h2>
                <p>{selectedRow.client.client_short_name}</p>
              </div>
              <button className="icon-button" type="button" onClick={closeEditor} aria-label="ปิดหน้าต่าง" disabled={saving}>×</button>
            </div>
            <form onSubmit={(event) => void saveService(event)}>
              <div className="form-grid">
                <label className="form-field form-field--wide">
                  <span>ชื่อบริการ <b>*</b></span>
                  <input required maxLength={150} value={form.serv_name} onChange={(event) => updateFormField('serv_name', event.target.value)} />
                </label>
                <label className="form-field">
                  <span>สถานะ UAT <b>*</b></span>
                  <select required value={form.uat_status} onChange={(event) => updateFormField('uat_status', event.target.value)}>
                    {UAT_STATUSES.map((status) => <option key={status}>{status}</option>)}
                  </select>
                </label>
                <label className="form-field">
                  <span>Production <b>*</b></span>
                  <select required value={form.production_status} onChange={(event) => updateFormField('production_status', event.target.value)}>
                    {PRODUCTION_STATUSES.map((status) => <option key={status.value} value={status.value}>{status.label}</option>)}
                  </select>
                </label>
                <label className="form-field">
                  <span>ผู้ทดสอบ</span>
                  <input maxLength={150} value={form.tested_by} onChange={(event) => updateFormField('tested_by', event.target.value)} placeholder="ชื่อผู้ทดสอบ" />
                </label>
                <label className="form-field">
                  <span>POS OS (NSS)</span>
                  <input maxLength={100} value={form.pos_os_nss} onChange={(event) => updateFormField('pos_os_nss', event.target.value)} placeholder="เช่น Windows 11" />
                </label>
                <label className="form-field">
                  <span>Week</span>
                  <input maxLength={20} value={form.week} onChange={(event) => updateFormField('week', event.target.value)} placeholder="เช่น Week 1" />
                </label>
                <label className="form-field">
                  <span>วันที่ส่ง UAT</span>
                  <input
                    type="text"
                    inputMode="numeric"
                    maxLength={10}
                    pattern="([0-9]{2}/[0-9]{2}/[0-9]{4})?"
                    title="กรอกวันที่ในรูปแบบ dd/mm/yyyy เช่น 25/10/2026"
                    value={form.uat_submitted_date}
                    onChange={(event) => updateFormField('uat_submitted_date', event.target.value)}
                    placeholder="dd/mm/yyyy"
                  />
                </label>
                <label className="form-field">
                  <span>ประเภทบริการ</span>
                  <input maxLength={100} value={form.service_type} onChange={(event) => updateFormField('service_type', event.target.value)} placeholder="เช่น Bill payment, Insurance" />
                </label>
                <label className="form-field form-field--wide">
                  <span>ลิงก์ไฟล์ Testcase &amp; Result</span>
                  <input type="url" maxLength={2048} value={form.testcase_result_link} onChange={(event) => updateFormField('testcase_result_link', event.target.value)} placeholder="https://" />
                </label>
              </div>
              {saveError && <p className="form-error" role="alert">{saveError}</p>}
              <div className="modal-actions">
                <button className="button button--soft" type="button" onClick={closeEditor} disabled={saving}>ยกเลิก</button>
                <button className="button button--primary" type="submit" disabled={saving}>
                  {saving ? <><span className="button-spinner" />กำลังบันทึก...</> : 'บันทึกการเปลี่ยนแปลง'}
                </button>
              </div>
            </form>
          </section>
        </div>
      )}
    </main>
  )
}

function StatCard({
  label,
  value,
  detail,
  tone,
  icon,
}: {
  label: string
  value: number
  detail: string
  tone: string
  icon: ReactNode
}) {
  return (
    <article className={`stat-card stat-card--${tone}`}>
      <div className="stat-topline"><span>{label}</span><span className="stat-icon">{icon}</span></div>
      <strong>{value}</strong>
      <span className="stat-detail">{detail}</span>
    </article>
  )
}

function ProgressStatCard({
  waitingCount,
  workingCount,
}: {
  waitingCount: number
  workingCount: number
}) {
  return (
    <article className="stat-card stat-card--progress">
      <div className="stat-topline">
        <span>ทดสอบ</span>
        <span className="stat-icon"><ClockIcon /></span>
      </div>
      <div className="issue-breakdown progress-breakdown">
        <div className="issue-breakdown-item progress-breakdown-item--waiting">
          <strong>{waitingCount}</strong>
          <span>Waiting test</span>
        </div>
        <div className="issue-breakdown-item progress-breakdown-item--working">
          <strong>{workingCount}</strong>
          <span>Working</span>
        </div>
      </div>
    </article>
  )
}

function ReleaseStatCard({
  readyCount,
  productionCount,
}: {
  readyCount: number
  productionCount: number
}) {
  return (
    <article className="stat-card stat-card--release">
      <div className="stat-topline">
        <span>สถานะการขึ้น Production</span>
        <span className="stat-icon"><CheckIcon /></span>
      </div>
      <div className="issue-breakdown release-breakdown">
        <div className="issue-breakdown-item release-breakdown-item--ready">
          <strong>{readyCount}</strong>
          <span>Pass UAT · รอขึ้น Prod.</span>
        </div>
        <div className="issue-breakdown-item release-breakdown-item--production">
          <strong>{productionCount}</strong>
          <span>ขึ้น Prod. แล้ว</span>
        </div>
      </div>
    </article>
  )
}

function SortableHeader({
  label,
  sortKey,
  sort,
  onSort,
}: {
  label: string
  sortKey: SortKey
  sort: { key: SortKey; direction: SortDirection }
  onSort: (key: SortKey) => void
}) {
  const active = sort.key === sortKey
  return (
    <button
      className={`sort-button ${active ? 'sort-button--active' : ''}`}
      type="button"
      aria-label={`เรียงตาม ${label}${active ? (sort.direction === 'asc' ? ' จากน้อยไปมาก' : ' จากมากไปน้อย') : ''}`}
      aria-pressed={active}
      onClick={() => onSort(sortKey)}
    >
      {label}
      <span aria-hidden="true">{active ? (sort.direction === 'asc' ? '↑' : '↓') : '↕'}</span>
    </button>
  )
}

function compareText(a: string, b: string, direction: SortDirection) {
  const result = a.localeCompare(b, undefined, { numeric: true, sensitivity: 'base' })
  return direction === 'asc' ? result : -result
}

function IssueStatCard({
  blockCount,
  blockByDefectCount,
  defectCount,
}: {
  blockCount: number
  blockByDefectCount: number
  defectCount: number
}) {
  return (
    <article className="stat-card stat-card--issues">
      <div className="stat-topline">
        <span>อยู่ระหว่างแก้ไข</span>
        <span className="stat-icon"><AlertIcon /></span>
      </div>
      <div className="issue-breakdown">
        <div className="issue-breakdown-item issue-breakdown-item--block">
          <strong>{blockCount}</strong>
          <span>Block</span>
        </div>
        <div className="issue-breakdown-item issue-breakdown-item--by-defect">
          <strong>{blockByDefectCount}</strong>
          <span>Block by defect</span>
        </div>
        <div className="issue-breakdown-item issue-breakdown-item--defect">
          <strong>{defectCount}</strong>
          <span>Defect</span>
        </div>
      </div>
    </article>
  )
}

function ProductionBadge({ status }: { status: string }) {
  const tone = status === 'PROD' ? 'production-pill--green' : status === 'NOT_APPLICABLE' ? 'production-pill--blue' : 'production-pill--purple'
  return <span className={`production-pill ${tone}`}><span />{status === 'PROD' ? 'ขึ้นแล้ว' : status === 'NOT_APPLICABLE' ? 'ไม่เกี่ยวข้อง' : 'ยังไม่ขึ้น'}</span>
}

function formatDate(date: string | null) {
  if (!date) return <span className="muted-value">—</span>
  const parsed = new Date(`${date}T00:00:00`)
  if (Number.isNaN(parsed.getTime())) return date
  return new Intl.DateTimeFormat('th-TH', { day: 'numeric', month: 'short', year: 'numeric' }).format(parsed)
}

function SearchIcon() {
  return <svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><circle cx="10.8" cy="10.8" r="6.8" stroke="currentColor" strokeWidth="1.7" /><path d="m16 16 4.2 4.2" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" /></svg>
}

function RefreshIcon() {
  return <svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M20 7v5h-5M4 17v-5h5" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" /><path d="M5.7 9a7 7 0 0 1 11.7-2L20 12M4 12l2.6 5a7 7 0 0 0 11.7-2" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" /></svg>
}

function GridIcon() {
  return <svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><rect x="4" y="4" width="6" height="6" rx="1.5" stroke="currentColor" strokeWidth="1.7" /><rect x="14" y="4" width="6" height="6" rx="1.5" stroke="currentColor" strokeWidth="1.7" /><rect x="4" y="14" width="6" height="6" rx="1.5" stroke="currentColor" strokeWidth="1.7" /><rect x="14" y="14" width="6" height="6" rx="1.5" stroke="currentColor" strokeWidth="1.7" /></svg>
}

function ClockIcon() {
  return <svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><circle cx="12" cy="12" r="8" stroke="currentColor" strokeWidth="1.7" /><path d="M12 7.5v5l3 1.8" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" /></svg>
}

function AlertIcon() {
  return <svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="m10.3 4.8-7 12.1A1.4 1.4 0 0 0 4.5 19h15a1.4 1.4 0 0 0 1.2-2.1l-7-12.1a2 2 0 0 0-3.4 0Z" stroke="currentColor" strokeWidth="1.7" strokeLinejoin="round" /><path d="M12 9v4.2M12 16.5v.1" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" /></svg>
}

function CheckIcon() {
  return <svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="m5 12.5 4.4 4.3L19.5 7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" /></svg>
}

export default App
