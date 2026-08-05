<template>
  <div class="users-page">
    <header class="topbar">
      <div class="brand">
        <div class="brand-badge">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
            <path d="M8 5.14v13.72c0 .81.89 1.3 1.57.87l10.6-6.86a1.04 1.04 0 0 0 0-1.74L9.57 4.27A1.03 1.03 0 0 0 8 5.14z" />
          </svg>
        </div>
        <span class="brand-name">LVS</span>
      </div>
      <div class="actions">
        <router-link to="/gallery" class="back-link">← 返回广场</router-link>
        <button class="logout-btn" @click="onLogout">注销</button>
      </div>
    </header>

    <section class="head">
      <div>
        <h1>用户管理</h1>
        <p>管理登录账号，管理员账号不可删除</p>
      </div>
      <button class="add-btn" @click="openCreate">新增用户</button>
    </section>

    <div v-if="loading" class="status"><div class="spinner"></div></div>

    <div v-else class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>用户名</th>
            <th>昵称</th>
            <th>角色</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td class="col-id">{{ u.id }}</td>
            <td>{{ u.username }}</td>
            <td>{{ u.nickname || '-' }}</td>
            <td>
              <span class="badge" :class="u.role === 'admin' ? 'badge-admin' : 'badge-user'">
                {{ u.role === 'admin' ? '管理员' : '普通用户' }}
              </span>
            </td>
            <td class="col-date">{{ u.created_at }}</td>
            <td class="col-ops">
              <button class="op-btn" @click="openEdit(u)">编辑</button>
              <button class="op-btn danger" :disabled="u.role === 'admin'" @click="onDelete(u)">
                删除
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新增/编辑弹窗 -->
    <div v-if="showModal" class="modal-mask" @click.self="showModal = false">
      <div class="modal">
        <h3>{{ editing ? '编辑用户' : '新增用户' }}</h3>
        <div class="field">
          <label>用户名</label>
          <input v-model.trim="form.username" :disabled="!!editing" placeholder="登录用户名" />
        </div>
        <div class="field">
          <label>昵称</label>
          <input v-model.trim="form.nickname" placeholder="显示昵称" />
        </div>
        <div class="field">
          <label>密码</label>
          <input
            v-model="form.password"
            type="password"
            :placeholder="editing ? '留空则不修改密码' : '登录密码'"
          />
        </div>
        <p v-if="formError" class="error">{{ formError }}</p>
        <div class="modal-actions">
          <button class="btn cancel" @click="showModal = false">取消</button>
          <button class="btn ok" :disabled="saving" @click="onSave">
            {{ saving ? '保存中…' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { fetchUsers, createUser, updateUser, deleteUser, logout } from '../api'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()
const users = ref([])
const loading = ref(false)

const showModal = ref(false)
const editing = ref(null)
const saving = ref(false)
const formError = ref('')
const form = ref({ username: '', nickname: '', password: '' })

async function load() {
  loading.value = true
  try {
    users.value = await fetchUsers()
  } catch (e) {
    // 拦截器已处理
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  form.value = { username: '', nickname: '', password: '' }
  formError.value = ''
  showModal.value = true
}

function openEdit(u) {
  editing.value = u
  form.value = { username: u.username, nickname: u.nickname, password: '' }
  formError.value = ''
  showModal.value = true
}

async function onSave() {
  if (!editing.value && !form.value.username) {
    formError.value = '请输入用户名'
    return
  }
  if (!editing.value && !form.value.password) {
    formError.value = '请输入密码'
    return
  }
  saving.value = true
  formError.value = ''
  try {
    if (editing.value) {
      await updateUser(editing.value.id, {
        nickname: form.value.nickname,
        password: form.value.password,
      })
    } else {
      await createUser({
        username: form.value.username,
        nickname: form.value.nickname,
        password: form.value.password,
      })
    }
    showModal.value = false
    load()
  } catch (e) {
    formError.value = e.message || '保存失败'
  } finally {
    saving.value = false
  }
}

async function onDelete(u) {
  if (!window.confirm(`确定删除用户 "${u.username}" 吗？`)) return
  try {
    await deleteUser(u.id)
    load()
  } catch (e) {
    window.alert(e.message || '删除失败')
  }
}

async function onLogout() {
  try {
    await logout()
  } catch (e) {
    /* ignore */
  }
  userStore.logout()
  router.push('/login')
}

onMounted(load)
</script>

<style scoped>
.users-page {
  min-height: 100vh;
  animation: fadeUp 0.4s ease both;
}
.topbar {
  position: sticky;
  top: 0;
  z-index: 20;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 28px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(11, 12, 17, 0.72);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
}
.brand {
  display: flex;
  align-items: center;
  gap: 10px;
}
.brand-badge {
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 9px;
  background: var(--grad);
  color: #fff;
  box-shadow: 0 6px 18px rgba(124, 92, 255, 0.4);
}
.brand-name {
  font-size: 17px;
  font-weight: 800;
  letter-spacing: 4px;
  background: var(--grad);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
.actions {
  display: flex;
  align-items: center;
  gap: 12px;
}
.back-link {
  padding: 8px 16px;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-2);
  font-size: 13px;
  transition: all 0.2s;
}
.back-link:hover {
  color: var(--text);
  background: rgba(255, 255, 255, 0.1);
}
.logout-btn {
  padding: 8px 16px;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-2);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}
.logout-btn:hover {
  color: var(--text);
  background: rgba(248, 113, 113, 0.12);
  border-color: rgba(248, 113, 113, 0.4);
}

.head {
  max-width: 1100px;
  margin: 0 auto;
  padding: 36px 24px 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.head h1 {
  font-size: 28px;
  font-weight: 800;
  letter-spacing: 1px;
  background: linear-gradient(120deg, #fff 20%, #a78bfa 60%, #22d3ee 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
.head p {
  margin-top: 6px;
  font-size: 13px;
  color: var(--text-2);
}
.add-btn {
  padding: 10px 22px;
  border: none;
  border-radius: 10px;
  background: var(--grad);
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 10px 28px rgba(124, 92, 255, 0.35);
  transition: transform 0.15s, box-shadow 0.2s;
}
.add-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 14px 36px rgba(124, 92, 255, 0.5);
}

.status {
  display: flex;
  justify-content: center;
  padding: 90px 0;
}

.table-wrap {
  max-width: 1100px;
  margin: 20px auto 40px;
  padding: 0 24px;
  overflow-x: auto;
}
table {
  width: 100%;
  border-collapse: collapse;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 14px;
  overflow: hidden;
}
thead th {
  padding: 14px 18px;
  text-align: left;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 1px;
  color: var(--text-3);
  text-transform: uppercase;
  background: rgba(255, 255, 255, 0.03);
  border-bottom: 1px solid var(--border);
}
tbody td {
  padding: 13px 18px;
  font-size: 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
}
tbody tr:last-child td {
  border-bottom: none;
}
tbody tr:hover {
  background: rgba(255, 255, 255, 0.03);
}
.col-id {
  color: var(--text-3);
  font-variant-numeric: tabular-nums;
}
.col-date {
  color: var(--text-2);
  font-size: 13px;
}
.badge {
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
}
.badge-admin {
  background: rgba(124, 92, 255, 0.18);
  color: #c4b5fd;
}
.badge-user {
  background: rgba(34, 211, 238, 0.14);
  color: #67e8f9;
}
.col-ops {
  white-space: nowrap;
}
.op-btn {
  padding: 6px 14px;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.05);
  color: var(--text);
  font-size: 13px;
  cursor: pointer;
  margin-right: 8px;
  transition: all 0.2s;
}
.op-btn:hover:not(:disabled) {
  background: rgba(124, 92, 255, 0.18);
  border-color: rgba(124, 92, 255, 0.5);
}
.op-btn.danger:hover:not(:disabled) {
  background: rgba(248, 113, 113, 0.15);
  border-color: rgba(248, 113, 113, 0.5);
}
.op-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

/* 弹窗 */
.modal-mask {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(6px);
}
.modal {
  width: 400px;
  max-width: calc(100vw - 40px);
  padding: 30px 32px;
  border-radius: 18px;
  background: #1a1c26;
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 30px 90px rgba(0, 0, 0, 0.6);
  animation: fadeUp 0.25s ease both;
}
.modal h3 {
  margin-bottom: 22px;
  font-size: 18px;
  font-weight: 700;
}
.field {
  margin-bottom: 16px;
}
.field label {
  display: block;
  margin-bottom: 7px;
  font-size: 13px;
  color: var(--text-2);
}
.field input {
  width: 100%;
  padding: 11px 14px;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.06);
  color: var(--text);
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.field input:focus {
  border-color: rgba(124, 92, 255, 0.7);
  box-shadow: 0 0 0 4px rgba(124, 92, 255, 0.16);
}
.field input:disabled {
  opacity: 0.5;
}
.error {
  color: var(--danger);
  font-size: 13px;
  margin: 2px 0 12px;
}
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 8px;
}
.btn {
  padding: 10px 22px;
  border-radius: 10px;
  font-size: 14px;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s;
}
.btn.cancel {
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-2);
  border-color: rgba(255, 255, 255, 0.14);
}
.btn.cancel:hover {
  color: var(--text);
  background: rgba(255, 255, 255, 0.1);
}
.btn.ok {
  background: var(--grad);
  color: #fff;
  font-weight: 600;
}
.btn.ok:hover:not(:disabled) {
  filter: brightness(1.1);
}
.btn.ok:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
