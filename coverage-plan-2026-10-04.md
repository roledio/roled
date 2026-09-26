# Rencana coverage sampai 4 Oktober 2026

Targetnya adalah minimal 80% overall untuk `auth` dan `console` secara terpisah. Target internal selesai 3 Oktober; 4 Oktober dipakai untuk audit akhir, clean-checkout rehearsal, dan perbaikan CI. Pengguna mengizinkan beberapa commit per hari, jadi prioritasnya adalah deadline dan kualitas test, bukan jumlah file per commit.

## Baseline 26 September

- Auth: 4.034/8.974 statement = **44,95%** pada scope yang sama dengan CI (`make test` mengecualikan cmd, configs, constants, entities, errors, mocks, models, views, migrations, databases, types, dan testutil). Perlu sekitar 3.146 statement tambahan teruji untuk mencapai 80% pada denominator saat ini.
- Console: 6.783/12.992 lines/statements = **52,20%**; branches 65,62% dan functions 40,52%. Baseline dijalankan dengan source `console/src`; laporan lama yang menghitung `dist` tidak dipakai. Ada satu test timeout pada percobaan paralel; percobaan ulang dengan dua worker menghasilkan report.

## Milestone

| Tanggal | Auth | Console | Fokus |
|---|---:|---:|---|
| 27 Sep | 52% | 58% | Bekukan scope; handler client/user; console core services, utilities, hooks |
| 28 Sep | 58% | 64% | Auth project/member/client; console API services dan hooks |
| 29 Sep | 64% | 69% | Auth user/role/resource/password; console CRUD dan error states |
| 30 Sep | 70% | 74% | Auth MariaDB/Redis repositories; console settings/profile/resources |
| 1 Okt | 75% | 78% | Auth middleware/web/queue/pkg; console auth/routing/branding |
| 2 Okt | 80% | 80% | Tutup gap terbesar berdasarkan laporan terbaru |
| 3 Okt | 82% | 82% | Buffer, edge cases, stabilisasi, clean-checkout rehearsal |
| 4 Okt | >=80% | >=80% | Final lint, typecheck, build, test, coverage dan laporan |

Target harian bersifat minimum; jika denominator berubah, laporkan numerator/denominator dan gap aktual. Gunakan beberapa sesi (09.00, 15.00, 20.00 WIB) dan beberapa commit substantif bila diperlukan.

## Prioritas teknis

Auth dimulai dari `internal/handlers/api/client.go`, lalu gap terbesar services (`project/create_project.go`, member/client/user flows), repositories MariaDB/Redis, middleware/web handlers, queues, dan pkg yang masih rendah. Console diprioritaskan berdasarkan lines uncovered: services project/roles/users/clients/resources, halaman Profile/Settings/Project/ResourceDetails, hooks, lalu komponen UI yang masuk scope. Setiap test harus menguji success, validation, error propagation, authorization/isolation, dan side effects yang relevan.

## Quality gate

Sebelum setiap commit: test relevan; `golangci-lint run` dari `auth`; untuk console `npm run lint`, `npx tsc --noEmit`, dan test relevan. Verifikasi lint/test auth pada clean checkout dengan mockery v3.8.0 seperti CI, supaya mock lokal yang di-ignore Git tidak menyembunyikan error. Akhir setiap hari jalankan coverage penuh kedua proyek. Jangan push, mengubah identitas Git, membuat commit kosong, atau menurunkan scope/threshold untuk mengejar angka.

## Risiko

Rata-rata perlu sekitar 450 statement auth tambahan per hari selama tujuh hari implementasi dan console perlu menambah sekitar 3.600 statement. Jika milestone tertinggal, sesi berikutnya langsung dialihkan ke file dengan gap terbesar dan hasil diukur ulang; jangan menunggu 4 Oktober. GitHub Actions baru dapat dipastikan setelah pengguna push, jadi sebelum itu laporkan hasil clean-checkout lokal sebagai verifikasi setara CI.

## Catatan sinkronisasi automation

Sebelum jadwal automation berikutnya dijalankan, prompt automation wajib memuat milestone terbaru, baseline 26 September, target minimal 80% untuk auth dan console, izin beberapa commit per hari, serta quality gate lint/test/clean-checkout yang tercantum di dokumen ini. Jika prompt belum tersinkron, lakukan sinkronisasi terlebih dahulu dan jangan memulai batch pekerjaan berdasarkan milestone lama.
