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
        <el-button @click="$router.push('/gallery')">← 返回广场</el-button>
        <el-button @click="onLogout">注销</el-button>
      </div>
    </header>

    <section class="head">
      <div>
        <h1>用户管理</h1>
        <p>管理登录账号，管理员账号不可删除</p>
      </div>
      <el-button type="primary" :icon="Plus" @click="openCreate">新增用户</el-button>
    </section>

    <div v-loading="loading" class="table-wrap" element-loading-text="加载中…">
      <el-table :data="users" style="width: 100%">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="username" label="用户名" min-width="140" />
        <el-table-column prop="nickname" label="昵称" min-width="140">
          <template #default="{ row }">{{ row.nickname || '-' }}</template>
        </el-table-column>
        <el-table-column label="角色" width="110">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'primary' : 'info'" effect="dark">
              {{ row.role === 'admin' ? '管理员' : '普通用户' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="170" />
        <el-table-column label="操作" width="160">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button
              size="small"
              type="danger"
              plain
              :disabled="row.role === 'admin'"
              @click="onDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="showModal"
      :title="editing ? '编辑用户' : '新增用户'"
      width="420px"
      align-center
    >
      <el-form label-position="top">
        <el-form-item label="用户名">
          <el-input v-model.trim="form.username" :disabled="!!editing" placeholder="登录用户名" />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model.trim="form.nickname" placeholder="显示昵称" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            :placeholder="editing ? '留空则不修改密码' : '登录密码'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showModal = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { fetchUsers, createUser, updateUser, deleteUser, logout } from '../api'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()
const users = ref([])
const loading = ref(false)

const showModal = ref(false)
const editing = ref(null)
const saving = ref(false)
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
  showModal.value = true
}

function openEdit(u) {
  editing.value = u
  form.value = { username: u.username, nickname: u.nickname, password: '' }
  showModal.value = true
}

async function onSave() {
  if (!editing.value && !form.value.username) {
    ElMessage.warning('请输入用户名')
    return
  }
  if (!editing.value && !form.value.password) {
    ElMessage.warning('请输入密码')
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await updateUser(editing.value.id, {
        nickname: form.value.nickname,
        password: form.value.password,
      })
      ElMessage.success('用户已更新')
    } else {
      await createUser({
        username: form.value.username,
        nickname: form.value.nickname,
        password: form.value.password,
      })
      ElMessage.success('用户已创建')
    }
    showModal.value = false
    load()
  } catch (e) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function onDelete(u) {
  try {
    await ElMessageBox.confirm(`确定删除用户 "${u.username}" 吗？`, '提示', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch (e) {
    return // 取消
  }
  try {
    await deleteUser(u.id)
    ElMessage.success('用户已删除')
    load()
  } catch (e) {
    ElMessage.error(e.message || '删除失败')
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

.head {
  max-width: 1100px;
  margin: 0 auto;
  padding: 36px 24px 16px;
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

.table-wrap {
  max-width: 1100px;
  margin: 16px auto 40px;
  padding: 0 24px;
}
.table-wrap :deep(.el-table) {
  --el-table-border-color: var(--border);
  --el-table-header-bg-color: rgba(255, 255, 255, 0.03);
  --el-table-row-hover-bg-color: rgba(255, 255, 255, 0.04);
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid var(--border);
}
</style>
