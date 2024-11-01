#ifndef AppVersion
  #define AppVersion "1.0.4"
#endif

#define MyAppName "Cursor History"
#define MyAppPublisher "Your Company"
#define MyAppURL "https://your-website.com"
#define MyAppExeName "CursorHistory.exe"

[Setup]
; 基本信息
AppId={{B8D80C89-F1D4-4F1D-9047-B0EAA2A5FF2F}
AppName={#MyAppName}
AppVersion={#AppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={userappdata}\{#MyAppName}
DefaultGroupName={#MyAppName}
AllowNoIcons=yes
OutputDir=installer
OutputBaseFilename=CursorHistory_Setup_{#AppVersion}
SetupIconFile=logo.ico
Compression=lzma
SolidCompression=yes
WizardStyle=modern

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[CustomMessages]
english.AutoStartup=Auto start with Windows
english.OtherOptions=Other Options

[Tasks]
Name: "desktopicon"; Description: "Create a &desktop icon"; GroupDescription: "{cm:AdditionalIcons}"
Name: "startupicon"; Description: "{cm:AutoStartup}"; GroupDescription: "{cm:OtherOptions}"

[Files]
; 主程序
Source: "CursorHistory.exe"; DestDir: "{app}"; Flags: ignoreversion
; 其他必要文件
Source: "logo.ico"; DestDir: "{app}"; Flags: ignoreversion
Source: "assets\*"; DestDir: "{app}\assets"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Registry]
; 添加开机启动项
Root: HKCU; Subkey: "Software\Microsoft\Windows\CurrentVersion\Run"; ValueType: string; ValueName: "CursorHistory"; ValueData: """{app}\{#MyAppExeName}"" -autostart"; Flags: uninsdeletevalue; Tasks: startupicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent