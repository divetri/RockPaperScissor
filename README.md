# RockPaperScissor (Implementing Server-Sent Event)

## 📌 Deskripsi

Proyek ini adalah permainan **Batu Gunting Kertas** berbasis web yang memungkinkan dua pemain bermain secara real-time menggunakan **Server-Sent Events (SSE)**. Backend dikembangkan menggunakan **Golang** untuk menangani permainan dan komunikasi antara pemain.

## 🚀 Cara Menjalankan Proyek

1. **Clone repository**
   ```sh
   git clone https://github.com/divetri/RockPaperScissor.git
   cd repository
   ```
2. **Jalankan server**
   ```sh
   go run main.go
   ```
3. **Akses permainan di browser**
   ```
   http://localhost:3000
   ```

## 🛠 Teknologi yang Digunakan

- **Golang** - Backend
- **Server-Sent Events (SSE)** - Untuk komunikasi real-time
- **HTML, CSS, JavaScript** - Frontend dasar

## 🎮 Cara Bermain

1. **Masuk ke Game**: Dua pemain membuka halaman permainan. Pastikan dua pemain masuk di server yang sama.
2. **Pilih Opsi**: Pilih batu, gunting, atau kertas.
3. **Tunggu Lawan**: Setelah kedua pemain memilih, hasil pertandingan ditampilkan secara real-time.
4. **Main Lagi**: Permainan bisa diulang dengan memilih kembali.

![Screen Capture](./screen-capture.gif)
![Diagram dot savage](./diagram.svg)

