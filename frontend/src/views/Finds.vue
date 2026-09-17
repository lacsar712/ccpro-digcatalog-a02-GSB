<template>
  <div>
    <div class="toolbar">
      <div>
        <h2 class="page-title">出土文物</h2>
        <p class="page-sub">登记器物类型、材质、完整度与存放位置</p>
      </div>
      <button class="btn" @click="openCreate">新增文物</button>
    </div>

    <div class="card">
      <div class="filters">
        <label>
          探方筛选
          <select v-model="filterUnitId" @change="load">
            <option value="">全部探方</option>
            <option v-for="u in units" :key="u.id" :value="String(u.id)">
              {{ u.site?.name || '' }} / {{ u.code }}
            </option>
          </select>
        </label>
        <label>
          器物类型
          <select v-model="filterType" @change="load">
            <option value="">全部类型</option>
            <option v-for="t in artifactTypes" :key="t" :value="t">{{ t }}</option>
          </select>
        </label>
      </div>

      <table class="table">
        <thead>
          <tr>
            <th>登记号</th>
            <th>探方</th>
            <th>器物类型</th>
            <th>材质</th>
            <th>完整度</th>
            <th>坐标(cm)</th>
            <th>出土日期</th>
            <th>存放位置</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in list" :key="item.id">
            <td>{{ item.registerNo }}</td>
            <td>{{ item.unit?.code || '-' }}</td>
            <td><span class="tag">{{ item.artifactType }}</span></td>
            <td>{{ item.materialName || item.material?.name || '-' }}</td>
            <td>{{ item.completeness || '-' }}</td>
            <td>{{ formatCoord(item) }}</td>
            <td>{{ formatDate(item.findDate) }}</td>
            <td>{{ item.storageLoc || '-' }}</td>
            <td>
              <button class="btn secondary small" @click="openEdit(item)">编辑</button>
              <button class="btn danger small" @click="remove(item)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
      <p v-if="!list.length" class="page-sub">暂无数据</p>
      <p v-if="error" class="error">{{ error }}</p>
    </div>

    <div v-if="showModal" class="modal-mask" @click.self="showModal = false">
      <div class="modal">
        <h3>{{ form.id ? '编辑文物' : '新增文物' }}</h3>
        <div class="form-grid">
          <label>
            所属探方
            <select v-model.number="form.unitId">
              <option :value="0" disabled>请选择</option>
              <option v-for="u in units" :key="u.id" :value="u.id">
                {{ u.site?.name || '' }} / {{ u.code }}
              </option>
            </select>
          </label>
          <label>
            登记号
            <input v-model="form.registerNo" />
          </label>
          <label>
            器物类型
            <select v-model="form.artifactType">
              <option v-for="t in artifactTypes" :key="t" :value="t">{{ t }}</option>
            </select>
          </label>
          <label>
            材质
            <select v-model="form.materialId">
              <option :value="null">未指定</option>
              <option v-for="m in materials" :key="m.id" :value="m.id">{{ m.name }}</option>
            </select>
          </label>
          <label>
            完整度
            <select v-model="form.completeness">
              <option>完整</option>
              <option>残缺</option>
              <option>碎片</option>
            </select>
          </label>
          <label>
            出土日期
            <input v-model="form.findDate" type="date" />
          </label>
          <div class="full coord-group">
            <span class="coord-title">
              探方局部坐标（厘米，整数，非负）
              <em>三值要么都留空，要么全部填写；同一探方内坐标唯一</em>
            </span>
            <div class="coord-inputs">
              <label>
                x
                <input v-model="form.xCm" type="number" min="0" step="1" placeholder="留空=未测点" />
              </label>
              <label>
                y
                <input v-model="form.yCm" type="number" min="0" step="1" placeholder="留空=未测点" />
              </label>
              <label>
                z
                <input v-model="form.zCm" type="number" min="0" step="1" placeholder="埋深" />
              </label>
            </div>
          </div>
          <label class="full">
            存放位置
            <input v-model="form.storageLoc" />
          </label>
          <label class="full">
            描述
            <textarea v-model="form.description" />
          </label>
        </div>
        <p v-if="formError" class="error">{{ formError }}</p>
        <div class="modal-actions">
          <button class="btn secondary" @click="showModal = false">取消</button>
          <button class="btn" @click="save">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import api from '../api/http'

const route = useRoute()
const artifactTypes = ['陶片', '青铜器', '骨器', '玉器', '石器', '铁器', '其他']
const list = ref([])
const units = ref([])
const materials = ref([])
const filterUnitId = ref('')
const filterType = ref('')
const error = ref('')
const formError = ref('')
const showModal = ref(false)

const form = reactive({
  id: null,
  unitId: 0,
  materialId: null,
  registerNo: '',
  artifactType: '陶片',
  completeness: '完整',
  findDate: '',
  xCm: '',
  yCm: '',
  zCm: '',
  description: '',
  storageLoc: ''
})

function formatDate(v) {
  if (!v) return '-'
  return String(v).slice(0, 10)
}

function formatCoord(item) {
  if (item.xCm == null || item.yCm == null || item.zCm == null) return '未测点'
  return `(${item.xCm}, ${item.yCm}, ${item.zCm})`
}

// 解析坐标输入：空串→null（未测）；否则须为非负整数。
function parseCoord(v) {
  if (v === '' || v === null || v === undefined) return { ok: true, value: null }
  const n = Number(v)
  if (!Number.isInteger(n) || n < 0) {
    return { ok: false, value: null }
  }
  return { ok: true, value: n }
}

async function loadMeta() {
  const [u, m] = await Promise.all([api.get('/units'), api.get('/materials')])
  units.value = u.data
  materials.value = m.data
}

async function load() {
  error.value = ''
  try {
    const params = {}
    if (filterUnitId.value) params.unitId = filterUnitId.value
    if (filterType.value) params.artifactType = filterType.value
    const { data } = await api.get('/finds', { params })
    list.value = data
  } catch (e) {
    error.value = e.response?.data?.error || '加载失败'
  }
}

function openCreate() {
  Object.assign(form, {
    id: null,
    unitId: Number(filterUnitId.value) || units.value[0]?.id || 0,
    materialId: materials.value[0]?.id ?? null,
    registerNo: '',
    artifactType: '陶片',
    completeness: '完整',
    findDate: '',
    xCm: '',
    yCm: '',
    zCm: '',
    description: '',
    storageLoc: ''
  })
  formError.value = ''
  showModal.value = true
}

function openEdit(item) {
  Object.assign(form, {
    id: item.id,
    unitId: item.unitId,
    materialId: item.materialId,
    registerNo: item.registerNo,
    artifactType: item.artifactType,
    completeness: item.completeness || '完整',
    findDate: formatDate(item.findDate) === '-' ? '' : formatDate(item.findDate),
    xCm: item.xCm ?? '',
    yCm: item.yCm ?? '',
    zCm: item.zCm ?? '',
    description: item.description || '',
    storageLoc: item.storageLoc || ''
  })
  formError.value = ''
  showModal.value = true
}

async function save() {
  formError.value = ''
  const xs = parseCoord(form.xCm)
  const ys = parseCoord(form.yCm)
  const zs = parseCoord(form.zCm)
  if (!xs.ok || !ys.ok || !zs.ok) {
    formError.value = '坐标必须是非负整数（厘米）'
    return
  }
  const filled = xs.value !== null || ys.value !== null || zs.value !== null
  if (filled && (xs.value === null || ys.value === null || zs.value === null)) {
    formError.value = '坐标需 x、y、z 三个值同时填写，或全部留空'
    return
  }
  try {
    const payload = {
      unitId: form.unitId,
      materialId: form.materialId || null,
      registerNo: form.registerNo,
      artifactType: form.artifactType,
      completeness: form.completeness,
      findDate: form.findDate || null,
      xCm: xs.value,
      yCm: ys.value,
      zCm: zs.value,
      description: form.description,
      storageLoc: form.storageLoc
    }
    if (form.id) {
      await api.put(`/finds/${form.id}`, payload)
    } else {
      await api.post('/finds', payload)
    }
    showModal.value = false
    await load()
  } catch (e) {
    formError.value = e.response?.data?.error || '保存失败'
  }
}

async function remove(item) {
  if (!confirm(`确认删除文物「${item.registerNo}」？`)) return
  try {
    await api.delete(`/finds/${item.id}`)
    await load()
  } catch (e) {
    alert(e.response?.data?.error || '删除失败')
  }
}

onMounted(async () => {
  if (route.query.unitId) filterUnitId.value = String(route.query.unitId)
  await loadMeta()
  await load()
})
</script>

<style scoped>
.filters {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}

.filters label {
  min-width: 200px;
}

.coord-group {
  border: 1px dashed var(--border);
  border-radius: 10px;
  padding: 0.75rem;
}

.coord-title {
  color: var(--text);
  font-size: 0.9rem;
  font-weight: 600;
}

.coord-title em {
  display: block;
  color: var(--muted);
  font-style: normal;
  font-weight: 400;
  font-size: 0.8rem;
  margin-top: 0.15rem;
}

.coord-inputs {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.75rem;
  margin-top: 0.6rem;
}
</style>
