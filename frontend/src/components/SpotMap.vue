<template>
  <div class="spotmap">
    <div class="spotmap-head">
      <span>出土点示意</span>
      <button class="refresh" title="刷新" @click="loadMap">↻</button>
    </div>
    <select v-model="unitId" class="spotmap-select" @change="loadMap">
      <option v-for="u in units" :key="u.id" :value="u.id">
        {{ u.site?.name || '' }} / {{ u.code }}
      </option>
    </select>
    <p v-if="error" class="spotmap-hint">{{ error }}</p>
    <template v-else-if="map">
      <p v-if="!configured" class="spotmap-hint">探方尺寸未配置，按 1000×1000 cm 示意</p>
      <svg :viewBox="`0 0 ${VB_W} ${VB_H}`" class="spotmap-svg">
        <rect :x="padL" :y="padT" :width="plotW" :height="plotH" class="grid-frame" />
        <line
          v-for="i in 4"
          :key="'v' + i"
          :x1="padL + (plotW * i) / 5"
          :y1="padT"
          :x2="padL + (plotW * i) / 5"
          :y2="padT + plotH"
          class="grid-line"
        />
        <line
          v-for="i in 4"
          :key="'h' + i"
          :x1="padL"
          :y1="padT + (plotH * i) / 5"
          :x2="padL + plotW"
          :y2="padT + (plotH * i) / 5"
          class="grid-line"
        />
        <text :x="padL" :y="padT + plotH + 12" class="axis-label" text-anchor="start">0</text>
        <text :x="padL + plotW" :y="padT + plotH + 12" class="axis-label" text-anchor="end">{{ effL }}</text>
        <text :x="padL - 4" :y="padT + plotH + 3" class="axis-label" text-anchor="end">0</text>
        <text :x="padL - 4" :y="padT + 7" class="axis-label" text-anchor="end">{{ effW }}</text>
        <circle
          v-for="s in spots"
          :key="s.id"
          :cx="px(s.xCm)"
          :cy="py(s.yCm)"
          :r="selected?.id === s.id ? 5.5 : 4"
          class="spot"
          :class="{ active: selected?.id === s.id }"
          @click="selectSpot(s)"
        >
          <title>{{ s.registerNo }}（{{ s.xCm }}, {{ s.yCm }}, {{ s.zCm }}）</title>
        </circle>
      </svg>
      <div class="spotmap-axis">X/cm → · Y/cm ↑（探方局部坐标）</div>
      <div v-if="selected" class="spotmap-info">
        <router-link :to="{ name: 'finds', query: { unitId } }" class="spotmap-reg">
          {{ selected.registerNo }}
        </router-link>
        <span class="spotmap-coord">x {{ selected.xCm }} · y {{ selected.yCm }} · z {{ selected.zCm }} cm</span>
      </div>
      <p v-else-if="!spots.length" class="spotmap-hint">该探方暂无带坐标的出土点</p>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import api from '../api/http'

const VB_W = 208
const VB_H = 196
const padL = 30
const padT = 8
const plotW = 168
const plotH = 164

const units = ref([])
const unitId = ref(null)
const map = ref(null)
const selected = ref(null)
const error = ref('')

const configured = computed(() => map.value?.unit?.lengthCm != null && map.value?.unit?.widthCm != null)
const spots = computed(() => map.value?.finds || [])
// 未配置尺寸时用 1000×1000 占位；坐标超出尺寸时扩展量程保证点可见
const effL = computed(() => Math.max(map.value?.unit?.lengthCm ?? 1000, ...spots.value.map((s) => s.xCm), 1))
const effW = computed(() => Math.max(map.value?.unit?.widthCm ?? 1000, ...spots.value.map((s) => s.yCm), 1))

function px(x) {
  return padL + (x / effL.value) * plotW
}

function py(y) {
  return padT + plotH - (y / effW.value) * plotH
}

async function loadUnits() {
  const { data } = await api.get('/units')
  units.value = data
  if (data.length && !unitId.value) {
    unitId.value = data[0].id
  }
}

async function loadMap() {
  if (!unitId.value) return
  error.value = ''
  selected.value = null
  try {
    const { data } = await api.get(`/units/${unitId.value}/spot-map`)
    map.value = data
  } catch (e) {
    map.value = null
    error.value = e.response?.data?.error || '加载失败'
  }
}

function selectSpot(s) {
  selected.value = selected.value?.id === s.id ? null : s
}

onMounted(async () => {
  try {
    await loadUnits()
    // 优先选中首个带坐标出土点的探方，都没有则回退到列表第一个
    for (const u of units.value) {
      const { data } = await api.get(`/units/${u.id}/spot-map`)
      if ((data.finds || []).length > 0) {
        unitId.value = u.id
        map.value = data
        return
      }
    }
    await loadMap()
  } catch (e) {
    error.value = e.response?.data?.error || '加载失败'
  }
})
</script>

<style scoped>
.spotmap {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.75rem 0.5rem 0.25rem;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.spotmap-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
  opacity: 0.85;
}

.refresh {
  background: none;
  border: none;
  color: inherit;
  cursor: pointer;
  opacity: 0.7;
  font-size: 0.95rem;
  padding: 0 0.2rem;
}

.refresh:hover {
  opacity: 1;
}

.spotmap-select {
  background: rgba(255, 255, 255, 0.08);
  color: var(--sidebar-text);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 8px;
  padding: 0.4rem 0.5rem;
  font-size: 0.8rem;
}

.spotmap-select option {
  color: #2a2218;
}

.spotmap-svg {
  width: 100%;
  height: auto;
  display: block;
}

.grid-frame {
  fill: rgba(196, 165, 116, 0.06);
  stroke: rgba(232, 223, 208, 0.35);
  stroke-width: 1;
}

.grid-line {
  stroke: rgba(232, 223, 208, 0.15);
  stroke-width: 1;
}

.axis-label {
  fill: rgba(232, 223, 208, 0.6);
  font-size: 8px;
}

.spot {
  fill: #c4a574;
  cursor: pointer;
  stroke: rgba(44, 36, 25, 0.6);
  stroke-width: 1;
}

.spot:hover {
  fill: #e0c391;
}

.spot.active {
  fill: #fff2d9;
  stroke: #c4a574;
}

.spotmap-axis {
  font-size: 0.7rem;
  opacity: 0.55;
  text-align: center;
}

.spotmap-hint {
  font-size: 0.75rem;
  opacity: 0.65;
  margin: 0;
}

.spotmap-info {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  background: rgba(196, 165, 116, 0.14);
  border-radius: 8px;
  padding: 0.45rem 0.6rem;
}

.spotmap-reg {
  font-weight: 700;
  font-size: 0.85rem;
  color: #e8c98f;
}

.spotmap-reg:hover {
  text-decoration: underline;
}

.spotmap-coord {
  font-size: 0.75rem;
  opacity: 0.75;
}

@media (max-width: 900px) {
  .spotmap {
    display: none;
  }
}
</style>
