# Apple Music Downloader (SaltPlayer Lyrics 适配版)

在原版基础上，适配了 **椒盐音乐** 的 **逐字歌词（含翻译）**，支持嵌入与独立导出。  

## ✨ 功能特色
- 🎵 支持 **逐字歌词 + 翻译**（中英文同步展示）  
- 📥 两种下载方式：
  1. **正常下载**：音频文件自动嵌入歌词  
  2. **仅歌词下载**：使用 `--lrc 专辑链接` 参数，保存为 `.lrc` 文件  
- 🌐 支持英文逐字 / 英文逐行 / 中文歌词  
- 📝 已对繁体歌词做适配，输出为简体中文  

---

## 📸 效果展示

### 英文（逐字）+ 翻译
![英文逐字+翻译](https://github.com/user-attachments/assets/d0a10543-ad54-447e-9db4-e55e406c8901)

### 英文（逐行）+ 翻译
![英文逐行+翻译](https://github.com/user-attachments/assets/9bfcdf02-3aa0-48c9-9996-26183c282e28)

### 英文 + 无翻译
（示例待补充）

### 中文（已简体化）
![中文简体](https://github.com/user-attachments/assets/b06e9baf-7b93-4cdb-bff3-bb01803da894)

---

## 🚀 使用方法
```bash
# 正常下载，自动嵌入歌词
与原版无差异，保留原版所有功能

# 仅下载歌词（保存为 .lrc）
./apple-music-downloader go run main.go --lrc <专辑链接>
