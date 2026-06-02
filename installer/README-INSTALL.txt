SIRUS FRISTA AGENT — CARA INSTALL
==================================

Agent ini membuka aplikasi FRISTA (Face Recognition BPJS), login otomatis,
dan mengetik nomor BPJS peserta — dipicu dari tombol "Scan Wajah" di web Sirus.

PENTING: agent ini JALAN DI SESI DESKTOP USER (autostart saat login),
BUKAN Windows Service. Karena harus mengetik ke jendela FRISTA, ia butuh
akses ke desktop yang sedang login. Jadi TIDAK perlu Run as administrator.

LANGKAH INSTALL
---------------
1. Siapkan file di PC pendaftaran — pilih salah satu:

   OPSI 1 (dari folder):
     Copy seluruh isi folder ini ke satu folder, mis. C:\SirusFristaAgent.

   OPSI 2 (dari file ZIP):
     a. Buat folder kosong dulu, mis. C:\SirusFristaAgent.
     b. Klik kanan ZIP -> Extract All... -> arahkan ke folder itu.
        (Isi ZIP ada di root, jadi WAJIB diekstrak ke folder khusus
         supaya 7 file-nya tidak berserakan.)

2. Buka config.json, isi:
     - "fristaPath" : lokasi frista.exe (default C:\frista\frista.exe)
     - "username"   : username login FRISTA
     - "password"   : password login FRISTA
3. Dobel-klik setup.bat (TANPA admin).
   -> file dicopy ke %LOCALAPPDATA%\SirusFristaAgent
   -> didaftarkan AUTOSTART (jalan tiap user login / setelah PC restart)
   -> agent langsung jalan, browser dashboard terbuka
4. Selesai. Coba dari dashboard: isi ID BPJS, klik "Buka FRISTA Sekarang".

CATATAN AUTOSTART
-----------------
Autostart BUKAN langkah terpisah. Ia aktif otomatis saat setup.bat dijalankan.
Ekstrak ZIP saja (tanpa setup.bat) -> agent BELUM autostart.
Cek aktif: HKCU\Software\Microsoft\Windows\CurrentVersion\Run -> entri
"SirusFristaAgent".

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
