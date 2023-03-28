package trayicon

type TrayIcon struct{}

func (icon *TrayIcon) Name() string {
	return "TrayIcon"
}

func (icon *TrayIcon) Content() []byte {
	return home_assistant_icon
}
