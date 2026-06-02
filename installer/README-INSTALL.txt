SIRUS FRISTA AGENT — CARA INSTALL
==================================

Agent ini membuka aplikasi FRISTA (Face Recognition BPJS), login otomatis,
dan mengetik nomor BPJS peserta — dipicu dari tombol "Scan Wajah" di web Sirus.

PENTING: agent ini JALAN DI SESI DESKTOP USER (autostart saat login),
BUKAN Windows Service. Karena harus mengetik ke jendela FRISTA, ia butuh
akses ke desktop yang sedang login. Jadi TIDAK perlu Run as administrator.

LANGKAH INSTALL
---------------
1. Copy seluruh isi folder ini ke PC pendaftaran (mis. C:\SirusFristaAgent).
2. Buka config.json, isi:
     - "fristaPath" : lokasi frista.exe (default C:\frista\frista.exe)
     - "username"   : username login FRISTA
     - "password"   : password login FRISTA
3. Dobel-klik setup.bat (TANPA admin).
   -> file dicopy ke %LOCALAPPDATA%\SirusFristaAgent
   -> didaftarkan autostart (jalan tiap user login)
   -> agent langsung jalan, browser dashboard terbuka
4. Selesai. Coba dari dashboard: isi ID BPJS, klik "Buka FRISTA Sekarang".

UPDATE
------
Jalankan setup.bat lagi. config.json yang sudah ada TIDAK ditimpa.

UNINSTALL
---------
Dobel-klik uninstall.bat.

TROUBLESHOOT
------------
- Agent tidak jalan setelah login? Cek HKCU\Software\Microsoft\Windows\
  CurrentVersion\Run ada entri "SirusFristaAgent".
- FRISTA terbuka tapi tidak terisi? Judul jendela mungkin beda dengan config
  "loginWindowTitle"/"mainWindowTitle". Sesuaikan, naikkan "stepDelayMs".
- Web bilang "agent tidak aktif"? Pastikan port di config.json = port di JS web,
  dan origin web ada di "allowedOrigins".
