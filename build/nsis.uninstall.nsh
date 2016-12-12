Section "Uninstall"
  # uninstall for all users
  setShellVarContext all

  # Delete (optionally) installed files
  {{range $}}Delete $INSTDIR\{{.}}
  {{end}}
  Delete $INSTDIR\uninstall.exe

  # Delete install directory
  rmDir $INSTDIR

  # Delete start menu launcher
  Delete "$SMPROGRAMS\${APPNAME}\${APPNAME}.lnk"
  Delete "$SMPROGRAMS\${APPNAME}\Attach.lnk"
  Delete "$SMPROGRAMS\${APPNAME}\Uninstall.lnk"
  rmDir "$SMPROGRAMS\${APPNAME}"

  # Firewall - remove rules if exists
  SimpleFC::AdvRemoveRule "Gur incoming peers (TCP:19595)"
  SimpleFC::AdvRemoveRule "Gur outgoing peers (TCP:19595)"
  SimpleFC::AdvRemoveRule "Gur UDP discovery (UDP:19595)"

  # Remove IPC endpoint (https://github.com/ethereum/EIPs/issues/147)
  ${un.EnvVarUpdate} $0 "ETHEREUM_SOCKET" "R" "HKLM" "\\.\pipe\gur.ipc"

  # Remove install directory from PATH
  ${un.EnvVarUpdate} $0 "PATH" "R" "HKLM" $INSTDIR

  # Cleanup registry (deletes all sub keys)
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${GROUPNAME} ${APPNAME}"
SectionEnd
