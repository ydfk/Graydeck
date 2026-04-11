import { create } from 'zustand'

type UiStore = {
  zashboardFrameMode: 'embedded' | 'fullscreen'
  setZashboardFrameMode: (mode: 'embedded' | 'fullscreen') => void
}

export const useUiStore = create<UiStore>((set) => ({
  zashboardFrameMode: 'embedded',
  setZashboardFrameMode: (mode) => set({ zashboardFrameMode: mode }),
}))
