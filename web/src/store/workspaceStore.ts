import { create } from 'zustand'
import type { Workspace } from '@/lib/api'

interface WorkspaceState {
  activeWorkspace: Workspace | null
  setActiveWorkspace: (w: Workspace | null) => void
}

export const useWorkspaceStore = create<WorkspaceState>()((set) => ({
  activeWorkspace: null,
  setActiveWorkspace: (w) => set({ activeWorkspace: w }),
}))
