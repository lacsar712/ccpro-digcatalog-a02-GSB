<template>
  <div>
    <div class="toolbar">
      <div>
        <h2 class="page-title">出土点示意</h2>
        <p class="page-sub">按探方局部坐标（厘米）展示文物出土平面位置，点击点位查看登记号</p>
      </div>
    </div>

    <div class="card">
      <label class="unit-select">
        选择探方
        <select v-model="unitId" @change="loadMap">
          <option value="" disabled>请选择探方</option>
          <option v-for="u in units" :key="u.id" :value="String(u.id)">
            {{ u.site?.name || '' }} / {{ u.code }}
          </option>
        </select>
      </label>
      <p v-if="error" class="error">{{ error }}</p>
    </div>

    <div v-if="map" class="board">
      <div class="card map-card">
        <div class="map-head">
          <div>
            <strong>{{ currentLabel }}</strong>
            <span class="dims">
              平面 {{ lengthCm }} × {{ widthCm }} cm
              <em v-if="!configured">（未配置尺寸，按 1000 × 1000 cm 占位）</em>
            </span>
          </div>
          <span class="tag">{{ map.spots.length }} 个已测点</span>
        </div>

        <div
          v-if="!configured"
          class="placeholder-tip"
        >
          该探方未配置长宽，图中为 1000 × 1000 cm 占位网格；请到「探方单位」补填实际尺寸。
        </div>

        <div class="svg-wrap" :style="{ aspectRatio: `${lengthCm} / ${widthCm}` }">
          <svg :viewBox="`0 0 ${lengthCm} ${widthCm}`" preserveAspectRatio="xMidYMid meet">
            <!-- 平面网格：每边 5 等分 -->
            <g class="grid">
              <line
                v-for="i in 4"
                :key="'vx' + i"
                :x1="(lengthCm / 5) * i"
                y1="0"
                :x2="(lengthCm / 5) * i"
                :y2="widthCm"
              />
              <line
                v-for="i in 4"
                :key="'hy' + i"
                x1="0"
                :y1="(widthCm / 5) * i"
                :x2="lengthCm"
                :y2="(widthCm / 5) * i"
              />
            </g>
            <rect x="0" y="0" :width="lengthCm" :height="widthCm" class="boundary" />

            <!-- 出土点 -->
            <g v-for="s in map.spots" :key="s.id">
              <circle
                class="spot"
                :class="{ active: selected && selected.id === s.id }"
                :cx="s.xCm"
                :cy="s.yCm"
                :r="dotR"
                @click="toggle(s)"
              />
              <text
                v-if="selected && selected.id === s.id"
                :x="s.xCm"
                :y="s.yCm - dotR * 1.6"
                class="spot-label"
                text-anchor="middle"
              >{{ s.registerNo }}</text>
            </g>
          </svg>

          <div v-if="selected" class="spot-pop" :style="popStyle" @click.stop>
            <div class="pop-no">
              <router-link :to="{ name: 'finds', query: { unitId: map.unitId } }">
                {{ selected.registerNo }}
              </router-link>
            </div>
            <div class="pop-type">{{ selected.artifactType || '文物' }}</div>
            <div class="pop-coord">
              x {{ selected.xCm }} · y {{ selected.yCm }} · z {{ selected.zCm }} cm
            </div>
            <button class="btn secondary small" @click="selected = null">关闭</button>
          </div>
        </div>

        <div class="axis-note">
          <span>← x 轴（长 {{ lengthCm }} cm）→</span>
          <span>↓ y 轴（宽 {{ widthCm }} cm），z 为埋深</span>
        </div>
      </div>

      <div class="card list-card">
        <h3>已测坐标文物</h3>
        <p v-if="!map.spots.length" class="page-sub">该探方暂无带坐标的文物</p>
        <ul class="spot-list">
          <li
            v-for="s in map.spots"
            :key="s.id"
            :class="{ active: selected && selected.id === s.id }"
            @click="toggle(s)"
          >
            <router-link
              class="no-link"
              :to="{ name: 'finds', query: { unitId: map.unitId } }"
              @click.stop
            >{{ s.registerNo }}</router-link>
            <span class="coord-text">({{ s.xCm }}, {{ s.yCm }}, {{ s.zCm }})</span>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import api from '../api/http'

const FALLBACK = 1000

const units = ref([])
const unitId = ref('')
const map = ref(null)
const selected = ref(null)
const error = ref('')

const configured = computed(
  () => !!(map.value && map.value.lengthCm && map.value.widthCm)
)
const lengthCm = computed(() => map.value?.lengthCm || FALLBACK)
const widthCm = computed(() => map.value?.widthCm || FALLBACK)
const dotR = computed(() => Math.max(lengthCm.value, widthCm.value) / 55)

const currentLabel = computed(() => {
  if (!map.value) return ''
  const u = units.value.find((x) => String(x.id) === String(unitId.value))
  return `${u?.site?.name || ''} / ${map.value.code}`
})

const popStyle = computed(() => {
  if (!selected.value) return {}
  return {
    left: `${(selected.value.xCm / lengthCm.value) * 100}%`,
    top: `${(selected.value.yCm / widthCm.value) * 100}%`
  }
})

async function loadUnits() {
  const { data } = await api.get('/units')
  units.value = data
  if (data.length) {
    unitId.value = String(data[0].id)
    await loadMap()
  }
}

async function loadMap() {
  error.value = ''
  selected.value = null
  map.value = null
  if (!unitId.value) return
  try {
    const { data } = await api.get(`/units/${unitId.value}/spot-map`)
    map.value = data
  } catch (e) {
    error.value = e.response?.data?.error || '加载示意图失败'
  }
}

function toggle(s) {
  selected.value = selected.value && selected.value.id === s.id ? null : s
}

onMounted(loadUnits)
</script>

<style scoped>
.unit-select {
  max-width: 320px;
}

.board {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 280px;
  gap: 1rem;
  margin-top: 1rem;
}

.map-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 0.75rem;
}

.dims {
  display: block;
  color: var(--muted);
  font-size: 0.85rem;
  margin-top: 0.2rem;
}

.dims em {
  color: var(--accent);
  font-style: normal;
}

.placeholder-tip {
  background: #f6ead9;
  border: 1px dashed var(--accent);
  color: var(--accent);
  border-radius: 8px;
  padding: 0.5rem 0.75rem;
  font-size: 0.85rem;
  margin-bottom: 0.75rem;
}

.svg-wrap {
  position: relative;
  width: 100%;
  border: 1px solid var(--border);
  border-radius: 10px;
  background:
    linear-gradient(0deg, rgba(139, 90, 43, 0.03), rgba(139, 90, 43, 0.03));
  overflow: hidden;
}

svg {
  display: block;
  width: 100%;
  height: 100%;
}

.grid line {
  stroke: #d9cbb8;
  stroke-width: 2;
  stroke-dasharray: 8 8;
}

.boundary {
  fill: none;
  stroke: var(--accent);
  stroke-width: 6;
}

.spot {
  fill: var(--danger);
  stroke: #fff;
  stroke-width: 3;
  cursor: pointer;
  transition: fill 0.15s;
}

.spot:hover,
.spot.active {
  fill: var(--accent);
}

.spot-label {
  font-size: 46px;
  font-weight: 700;
  fill: var(--text);
  paint-order: stroke;
  stroke: #fffdf8;
  stroke-width: 10px;
}

.spot-pop {
  position: absolute;
  transform: translate(-50%, calc(-100% - 14px));
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 10px;
  box-shadow: var(--shadow);
  padding: 0.6rem 0.75rem;
  min-width: 150px;
  z-index: 5;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.pop-no {
  font-weight: 700;
}

.pop-no a:hover {
  color: var(--accent);
}

.pop-type {
  font-size: 0.85rem;
  color: var(--accent);
}

.pop-coord {
  font-size: 0.82rem;
  color: var(--muted);
}

.axis-note {
  display: flex;
  justify-content: space-between;
  margin-top: 0.5rem;
  color: var(--muted);
  font-size: 0.82rem;
}

.list-card h3 {
  margin: 0 0 0.75rem;
  font-size: 1rem;
}

.spot-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.spot-list li {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 0.5rem;
  padding: 0.45rem 0.6rem;
  border: 1px solid var(--border);
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.88rem;
}

.spot-list li.active {
  background: #f6ead9;
  border-color: var(--accent);
}

.no-link {
  font-weight: 600;
}

.no-link:hover {
  color: var(--accent);
}

.coord-text {
  color: var(--muted);
  font-size: 0.8rem;
}

@media (max-width: 900px) {
  .board {
    grid-template-columns: 1fr;
  }
}
</style>
