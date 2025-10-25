// api.js — AlgoCDK Frontend API Handler

const API_BASE_URL = "http://localhost:8080"; // change to your deployed domain later

// Helper: Make requests with auth header if token exists
async function apiRequest(endpoint, method = "GET", data = null, isForm = false) {
  const token = localStorage.getItem("token");
  const headers = token ? { "Authorization": `Bearer ${token}` } : {};

  if (!isForm && data) headers["Content-Type"] = "application/json";

  const options = {
    method,
    headers,
  };

  if (data) {
    options.body = isForm ? data : JSON.stringify(data);
  }

  const res = await fetch(`${API_BASE_URL}${endpoint}`, options);
  if (!res.ok) throw new Error(`Request failed: ${res.status}`);
  return await res.json();
}

/* =====================
   AUTH ENDPOINTS
===================== */
export async function login(email, password) {
  return apiRequest("/api/auth/login", "POST", { email, password });
}

export async function register(userData) {
  return apiRequest("/api/auth/register", "POST", userData);
}

/* =====================
   USER ACTIONS
===================== */
export async function getProfile() {
  return apiRequest("/api/user/me");
}

export async function getFavorites() {
  return apiRequest("/api/user/me/favorites");
}

export async function toggleFavorite(botId) {
  return apiRequest(`/api/user/favorites/${botId}`, "POST");
}

export async function requestUpgrade() {
  return apiRequest("/api/user/request-upgrade", "POST");
}

/* =====================
   MARKETPLACE & BOTS
===================== */
export async function getMarketplace() {
  return apiRequest("/marketplace");
}

export async function getAdminBots() {
  return apiRequest("/api/admin/bots");
}

export async function createBot(botData) {
  return apiRequest("/api/admin/create-bot", "POST", botData);
}

export async function updateBot(botId, botData) {
  return apiRequest(`/api/admin/update-bot/${botId}`, "PUT", botData);
}

export async function deleteBot(botId) {
  return apiRequest(`/api/admin/delete-bot/${botId}`, "DELETE");
}

/* =====================
   ADMIN TRANSACTIONS
===================== */
export async function getAdminTransactions() {
  return apiRequest("/api/admin/transactions");
}

export async function recordTransaction(transactionData) {
  return apiRequest("/api/admin/transactions", "POST", transactionData);
}

/* =====================
   SUPERADMIN ACTIONS
===================== */
export async function superAdminLogin(credentials) {
  return apiRequest("/api/superadmin/login", "POST", credentials);
}

export async function superAdminRegister(data) {
  return apiRequest("/api/superadmin/register", "POST", data);
}

export async function getSuperAdminProfile() {
  return apiRequest("/api/superadmin/profile");
}

export async function getAllUsers() {
  return apiRequest("/api/superadmin/users");
}

export async function createUser(userData) {
  return apiRequest("/api/superadmin/create-user", "POST", userData);
}

export async function updateUser(userData) {
  return apiRequest("/api/superadmin/update-user", "POST", userData);
}

export async function deleteUser(userId) {
  return apiRequest("/api/superadmin/delete-user", "DELETE", { id: userId });
}

export async function getPendingRequests() {
  return apiRequest("/api/superadmin/pending-requests");
}

export async function approveUpgrade(userId) {
  return apiRequest(`/api/superadmin/promote/${userId}`, "POST");
}

export async function rejectUpgrade(userId) {
  return apiRequest(`/api/superadmin/reject/${userId}`, "POST");
}

export async function getAllAdmins() {
  return apiRequest("/api/superadmin/admins");
}

export async function createAdmin(adminData) {
  return apiRequest("/api/superadmin/create-admin", "POST", adminData);
}

export async function updateAdmin(adminId, adminData) {
  return apiRequest(`/api/superadmin/update-admin/${adminId}`, "PUT", adminData);
}

export async function toggleAdminStatus(adminId) {
  return apiRequest(`/api/superadmin/toggle-admin/${adminId}`, "PATCH");
}

export async function deleteAdmin(adminId) {
  return apiRequest(`/api/superadmin/delete-admin/${adminId}`, "DELETE");
}

/* =====================
   SUPERADMIN BOTS
===================== */
export async function scanAllBots() {
  return apiRequest("/api/superadmin/scan-bots", "POST"); // or "GET", both will work now
}

export async function getAllTransactions() {
  return apiRequest("/api/superadmin/transactions", "GET");
}

/* =====================
   PAYSTACK INTEGRATION
===================== */
export async function initializePayment(data) {
  return apiRequest("/api/user/paystack/init", "POST", data);
}

export async function verifyPayment(reference) {
  return apiRequest(`/api/user/paystack/verify?reference=${reference}`);
}

/* =====================
   DASHBOARDS
===================== */
export async function getAdminDashboard() {
  return apiRequest("/api/admin/dashboard");
}

export async function getSuperAdminDashboard() {
  return apiRequest("/api/superadmin/dashboard");
}

/* =====================
   WEBSOCKET
===================== */
export function connectWebSocket(role = "user") {
  let url;
  if (role === "superadmin") url = `${API_BASE_URL}/api/superadmin/ws`;
  else if (role === "user") url = `${API_BASE_URL}/api/user/ws`;
  else url = `${API_BASE_URL}/api/admin/ws`;
  return new WebSocket(url);
}