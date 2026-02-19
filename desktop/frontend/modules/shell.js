export const createShellController = ({ bridge, refs, readSelection, onStatus, onToast }) => {
  const loadShellUser = async () => {
    const project = readSelection().project
    if (!project || !refs.shellUser) {
      return
    }
    try {
      const user = await bridge.getShellUser(project)
      refs.shellUser.value = user || ""
    } catch (_err) {
      refs.shellUser.value = ""
    }
  }

  const saveShellUser = async () => {
    const project = readSelection().project
    if (!project || !refs.shellUser) {
      return
    }
    try {
      await bridge.setShellUser(project, refs.shellUser.value || "")
    } catch (_err) {
      // Keep UI non-blocking for preference persistence.
    }
  }

  const openShell = async () => {
    const { project, service } = readSelection()
    if (!project) {
      onStatus("Select an environment to open shell.")
      return
    }
    const shellUser = refs.shellUser?.value || ""
    const shell = refs.shellCommand?.value || "bash"
    try {
      const message = await bridge.openShellForService(project, service, shellUser, shell)
      onStatus(message)
      onToast(message, "success")
    } catch (err) {
      const message = `Failed to open shell: ${err}`
      onStatus(message)
      onToast(message, "error")
    }
  }

  const resetShellUsers = async () => {
    try {
      const message = await bridge.resetShellUsers()
      onStatus(message)
      onToast(message, "success")
      await loadShellUser()
    } catch (err) {
      const message = `Failed to reset shell users: ${err}`
      onStatus(message)
      onToast(message, "error")
    }
  }

  return { loadShellUser, saveShellUser, openShell, resetShellUsers }
}

