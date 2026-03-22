import { getAccessToken } from "@/shared/auth/token-storage"
import type { CreateClientConfig } from "./generated/client.gen"

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  baseUrl: "/api/v1",
  credentials: "include",
  auth: () => getAccessToken() ?? "",
})
