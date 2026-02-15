package db

import (
	"context"
	"database/sql"
	"log"
	"math/rand"
	"socialv3/internal/store"
)

var usernames = []string{
	"abc", "aaa", "bbb", "ccc", "ddd", "eee", "fff", "def", "wow", "red", "mame", "tod", "sam",
	"randi", "wordy", "lower", "upper", "elephant", "ranter", "someter", "tingker", "langer",
	"brute", "rnnd", "rhino", "someth", "thinger", "rjh", "roro", "morty", "angel", "deomon",
	"rust", "java", "script", "python", "snake", "yasmin", "odin", "thor", "anderson", "morror",
	"windows", "ubuntu", "linux", "maxcy", "max", "minus", "tower",
}

var paragraph = []string{
	"Langkah Kecil yang Tanpa Disadari Mengubah Arah Hidup Seseorang",
	"Malam Panjang Ketika Semua Pertanyaan Datang Tanpa Jawaban",
	"Cerita di Balik Layar yang Tidak Pernah Sampai ke Publik",
	"Jejak Waktu yang Tertinggal di Tempat yang Sudah Lama Ditinggalkan",
	"Antara Kopi yang Mendingin dan Hujan yang Tidak Kunjung Berhenti",
	"Suara Pelan dari Selatan yang Membawa Kenangan Lama",
	"Hari yang Terlewat Begitu Saja Tanpa Ada yang Benar-Benar Disadari",
	"Rencana Besar yang Harus Ditunda Karena Keadaan yang Tak Terduga",
	"Detik-Detik Sunyi Sebelum Pagi Datang Mengganti Segalanya",
	"Rahasia yang Tersembunyi di Ujung Jalan yang Jarang Dilalui Orang",
	"Nada Musik yang Hilang di Tengah Keramaian Kota",
	"Peta Perjalanan Tanpa Arah Jelas Namun Terus Dilanjutkan",
	"Senja Terakhir yang Terasa Lebih Lama dari Biasanya",
	"Surat Panjang yang Ditulis Namun Tak Pernah Berani Dikirim",
	"Bayangan Masa Lalu yang Kembali Muncul Saat Semua Terlihat Tenang",
	"Langit Cerah Setelah Hujan Lebat yang Mengubah Suasana Hati",
	"Percakapan Panjang di Tengah Malam yang Mengubah Cara Pandang",
	"Titik Balik Kehidupan yang Datang Tanpa Peringatan",
	"Satu Langkah Lagi Sebelum Segalanya Menjadi Berbeda",
	"Waktu yang Dipinjam untuk Menyelesaikan Hal yang Belum Sempat",
	"Nama Lama yang Kembali Disebut Setelah Bertahun-Tahun Terlupakan",
	"Perjalanan Panjang Menuju Jalan Pulang yang Sebenarnya",
	"Di Antara Dua Dunia yang Sama-Sama Tidak Sepenuhnya Nyata",
	"Hening Panjang yang Justru Menyampaikan Banyak Hal",
	"Catatan Kecil yang Mengubah Cara Seseorang Mengingat Masa Lalu",
	"Ruang Kosong yang Perlahan Terisi oleh Hal-Hal Tak Terduga",
	"Janji Lama di Bulan Juni yang Akhirnya Harus Dihadapi",
	"Sebelum Semuanya Berubah dan Tidak Bisa Kembali Seperti Semula",
	"Potongan Cerita Kecil yang Membentuk Gambaran Besar",
	"Arah yang Sama Namun Tujuan yang Sangat Berbeda",
	"Detak Jantung yang Terasa Berbeda Saat Kesadaran Datang",
	"Sisa Cahaya di Ujung Hari yang Hampir Terlewat",
	"Tanpa Banyak Kata Namun Sarat Akan Makna",
	"Hari Pertama yang Terasa Seperti Awal dari Segalanya",
	"Batas Waktu yang Terus Mendekat Tanpa Bisa Dihentikan",
	"Langkah Terakhir Sebelum Mengambil Keputusan Besar",
	"Pesan Singkat yang Mengubah Suasana Hati Sepanjang Hari",
	"Cerita yang Terlihat Biasa dari Balik Sebuah Jendela",
	"Garis Tipis Antara Bertahan dan Melepaskan",
	"Kembali ke Awal untuk Memahami Kesalahan yang Sama",
	"Cerita Sementara yang Ternyata Membekas Lebih Lama",
	"Jam yang Berhenti Tepat Saat Semuanya Terasa Penuh",
	"Satu Tanda Tanya Besar yang Tidak Pernah Terjawab",
	"Rencana Cadangan yang Akhirnya Menjadi Pilihan Utama",
	"Satu Malam Panjang yang Mengubah Banyak Hal Sekaligus",
	"Di Ujung Waktu Ketika Tidak Ada yang Bisa Dipaksakan",
	"Fragmen Kecil Kenangan yang Terus Muncul Tanpa Diundang",
	"Sisi Lain dari Cerita yang Selama Ini Disembunyikan",
	"Setelah Semua Usai dan Tidak Ada Lagi yang Bisa Diulang",
}

var tags = []string{
	"kehidupan",
	"cerita",
	"refleksi",
	"motivasi",
	"inspirasi",
	"pengalaman",
	"perjalanan",
	"waktu",
	"kenangan",
	"perubahan",
	"pilihan",
	"masa-depan",
	"masa-lalu",
	"hari-ini",
	"emosi",
	"pikiran",
	"kesadaran",
	"pertumbuhan",
	"tujuan",
	"harapan",
}

var alphaNum = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func randChar(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = alphaNum[rand.Intn(len(alphaNum))]
	}
	return string(b)
}

func generateUsers(num int) []*store.User {
	users := make([]*store.User, num)
	for i := 0; i < num; i++ {
		users[i] = &store.User{
			Username: usernames[rand.Intn(len(usernames))] + randChar(50),
			Email:    usernames[rand.Intn(len(usernames))] + randChar(50) + "@example.com",
		}
		if err := users[i].Password.Set("foobar12345"); err != nil {
			panic(err)
		}
	}
	return users
}

func generatePostsByUsers(num int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, num)
	for i := 0; i < num; i++ {
		posts[i] = &store.Post{
			Title:   paragraph[rand.Intn(len(paragraph))] + randChar(20),
			Content: paragraph[rand.Intn(len(paragraph))] + paragraph[rand.Intn(len(paragraph))] + paragraph[rand.Intn(len(paragraph))],
			UserID:  users[rand.Intn(len(users))].ID,
			Tags: []string{
				tags[rand.Intn(len(tags))],
				tags[rand.Intn(len(tags))],
			},
		}
	}
	return posts
}

func generateCommentsByUsersAndPosts(num int, users []*store.User, posts []*store.Post) []*store.Comment {
	comments := make([]*store.Comment, num)
	for i := 0; i < num; i++ {
		comments[i] = &store.Comment{
			PostID:  posts[i].ID,
			UserID:  users[i].ID,
			Content: posts[rand.Intn(len(posts))].Content,
		}
	}
	return comments
}

func Seed(db *sql.DB, store store.Storage) error {
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	users := generateUsers(100)
	for _, user := range users {
		if err := store.Users.Create(ctx, tx, user); err != nil {
			log.Println("error creating user", user.Username, err)
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	posts := generatePostsByUsers(500, users)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			log.Println("error creating post", post.Title, err)
			return err
		}
	}

	comments := generateCommentsByUsersAndPosts(100, users, posts)
	for _, comment := range comments {
		if err := store.Comments.Create(ctx, comment); err != nil {
			log.Println("error creating comment", comment.PostID, err)
			return err
		}
	}

	log.Println("Seeding complete")
	return nil
}
