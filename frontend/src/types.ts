export type Service = {
  client_service_id: string
  service_id: string
  serv_name: string
  uat_status: string
  tested_by: string | null
  pos_os_nss: string | null
  testcase_result_link: string | null
  week: string | null
  uat_submitted_date: string | null
  service_type: string | null
  production_status: string
}

export type Client = {
  client_id: string
  client_short_name: string
  services: Service[]
}

export type ServiceUpdate = {
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
