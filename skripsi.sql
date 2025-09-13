-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: 127.0.0.1
-- Generation Time: Sep 13, 2025 at 08:30 PM
-- Server version: 10.4.32-MariaDB
-- PHP Version: 8.2.12

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `skripsi`
--

-- --------------------------------------------------------

--
-- Table structure for table `akun`
--

CREATE TABLE `akun` (
  `akun_id` int(11) NOT NULL,
  `kode_akun` varchar(10) NOT NULL,
  `nama_akun` varchar(100) NOT NULL,
  `jenis` enum('Asset','Liability','Equity','Revenue','Expense') NOT NULL,
  `parent_id` int(11) DEFAULT NULL,
  `is_header` tinyint(1) DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `akun`
--

INSERT INTO `akun` (`akun_id`, `kode_akun`, `nama_akun`, `jenis`, `parent_id`, `is_header`) VALUES
(1, '1-000', 'Aset', 'Asset', NULL, 1),
(2, '1-100', 'Aset Lancar', 'Asset', 1, 1),
(3, '1-101', 'Kas', 'Asset', 2, 0),
(4, '1-102', 'Bank', 'Asset', 2, 0),
(5, '1-103', 'Piutang Usaha', 'Asset', 2, 0),
(6, '1-104', 'Persediaan', 'Asset', 2, 0),
(7, '1-200', 'Aset Tetap', 'Asset', 1, 1),
(8, '1-201', 'Peralatan', 'Asset', 7, 0),
(9, '1-202', 'Akumulasi Penyusutan', 'Asset', 7, 0),
(10, '2-000', 'Liabilitas', 'Liability', NULL, 1),
(11, '2-100', 'Liabilitas Jangka Pendek', 'Liability', 10, 1),
(12, '2-101', 'Utang Usaha', 'Liability', 11, 0),
(13, '2-102', 'Utang Pajak', 'Liability', 11, 0),
(14, '2-200', 'Liabilitas Jangka Panjang', 'Liability', 10, 1),
(15, '2-201', 'Pinjaman Bank', 'Liability', 14, 0),
(16, '3-000', 'Ekuitas', 'Equity', NULL, 1),
(17, '3-100', 'Modal Pemilik', 'Equity', 16, 0),
(18, '3-200', 'Laba Ditahan', 'Equity', 16, 0),
(19, '4-000', 'Pendapatan', 'Revenue', NULL, 1),
(20, '4-100', 'Pendapatan Usaha', 'Revenue', 19, 1),
(21, '4-101', 'Penjualan Produk', 'Revenue', 20, 0),
(22, '4-102', 'Pendapatan Lain', 'Revenue', 20, 0),
(23, '5-000', 'Beban', 'Expense', NULL, 1),
(24, '5-100', 'Harga Pokok Penjualan', 'Expense', 23, 0),
(25, '5-200', 'Beban Operasional', 'Expense', 23, 1),
(26, '5-201', 'Beban Gaji', 'Expense', 25, 0),
(27, '5-202', 'Beban Listrik dan Air', 'Expense', 25, 0),
(28, '5-203', 'Beban Sewa', 'Expense', 25, 0),
(29, '5-204', 'Beban Transportasi', 'Expense', 25, 0),
(30, '3-101', 'Prive', 'Equity', 16, 0);

-- --------------------------------------------------------

--
-- Table structure for table `barang`
--

CREATE TABLE `barang` (
  `barang_id` int(11) NOT NULL,
  `kode_barang` varchar(255) DEFAULT NULL,
  `nama_barang` varchar(255) DEFAULT NULL,
  `harga_jual` decimal(10,0) DEFAULT NULL,
  `harga_beli` decimal(10,0) DEFAULT NULL,
  `jumlah_stock` int(11) DEFAULT NULL,
  `is_active` tinyint(1) NOT NULL DEFAULT 1
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `barang`
--

INSERT INTO `barang` (`barang_id`, `kode_barang`, `nama_barang`, `harga_jual`, `harga_beli`, `jumlah_stock`, `is_active`) VALUES
(322, '35', 'Klt DIMSUM MB', 11000, 0, 20, 1),
(323, '99', 'MAK E PENTOL 50', 17800, 0, 20, 1),
(324, '8993200664399', 'KANZLER CRISPY CN BC', 41000, 0, 20, 1),
(325, '8993207730035', 'CHAMP SOSIS AYAM 15', 18000, 0, 20, 1),
(326, '8997014710044', 'GSTAR CN KEJU', 42500, 0, 20, 1),
(327, '8998888461124', 'MAESTRO Mayonais 180gr', 8500, 0, 20, 1),
(328, '8998888710192', 'DELM SAOS TOMAT 200GR', 6000, 0, 20, 1),
(329, '8998888710598', 'DELM EXTRA PEDAS 200GR', 6500, 0, 20, 1);

-- --------------------------------------------------------

--
-- Table structure for table `barang_keluar`
--

CREATE TABLE `barang_keluar` (
  `keluar_id` int(11) NOT NULL,
  `penjualan_id` int(11) DEFAULT NULL,
  `barang_id` int(11) DEFAULT NULL,
  `masuk_id` int(11) DEFAULT NULL,
  `jumlah` int(11) DEFAULT NULL,
  `harga_beli` decimal(10,0) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `barang_keluar`
--

INSERT INTO `barang_keluar` (`keluar_id`, `penjualan_id`, `barang_id`, `masuk_id`, `jumlah`, `harga_beli`) VALUES
(396, 189, 326, 4158, 1, 38000),
(397, 189, 322, 4154, 1, 10000),
(398, 190, 329, 4161, 1, 4500),
(399, 190, 328, 4160, 1, 4000),
(400, 190, 327, 4159, 1, 5500),
(401, 190, 325, 4157, 1, 10000),
(402, 190, 324, 4156, 1, 35000),
(403, 191, 326, 4158, 1, 38000),
(404, 191, 322, 4154, 1, 10000),
(405, 192, 329, 4161, 1, 4500),
(406, 192, 328, 4160, 1, 4000),
(407, 192, 327, 4159, 1, 5500),
(408, 192, 325, 4157, 1, 10000),
(409, 192, 324, 4156, 1, 35000);

-- --------------------------------------------------------

--
-- Table structure for table `barang_masuk`
--

CREATE TABLE `barang_masuk` (
  `masuk_id` int(11) NOT NULL,
  `barang_id` int(11) DEFAULT NULL,
  `tanggal` date DEFAULT NULL,
  `jumlah` int(11) DEFAULT NULL,
  `harga_beli` decimal(10,0) DEFAULT NULL,
  `sisa_stok` int(11) DEFAULT NULL,
  `pembelian_id` int(11) DEFAULT NULL,
  `keterangan` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `barang_masuk`
--

INSERT INTO `barang_masuk` (`masuk_id`, `barang_id`, `tanggal`, `jumlah`, `harga_beli`, `sisa_stok`, `pembelian_id`, `keterangan`) VALUES
(4154, 322, '2025-08-05', 10, 10000, 8, NULL, 'Upload CSV Baru'),
(4155, 323, '2025-08-05', 10, 15000, 10, NULL, 'Upload CSV Baru'),
(4156, 324, '2025-08-05', 10, 35000, 8, NULL, 'Upload CSV Baru'),
(4157, 325, '2025-08-05', 10, 10000, 8, NULL, 'Upload CSV Baru'),
(4158, 326, '2025-08-05', 10, 38000, 8, NULL, 'Upload CSV Baru'),
(4159, 327, '2025-08-05', 10, 5500, 8, NULL, 'Upload CSV Baru'),
(4160, 328, '2025-08-05', 10, 4000, 8, NULL, 'Upload CSV Baru'),
(4161, 329, '2025-08-05', 10, 4500, 8, NULL, 'Upload CSV Baru'),
(4162, 326, '2025-09-02', 3, 38000, 3, 12, NULL),
(4163, 322, '2025-09-05', 10, 10000, 10, NULL, 'Upload CSV Tambahan'),
(4164, 323, '2025-09-05', 10, 15000, 10, NULL, 'Upload CSV Tambahan'),
(4165, 324, '2025-09-05', 10, 35000, 10, NULL, 'Upload CSV Tambahan'),
(4166, 325, '2025-09-05', 10, 10000, 10, NULL, 'Upload CSV Tambahan'),
(4167, 326, '2025-09-05', 10, 38000, 10, NULL, 'Upload CSV Tambahan'),
(4168, 327, '2025-09-05', 10, 5500, 10, NULL, 'Upload CSV Tambahan'),
(4169, 328, '2025-09-05', 10, 4000, 10, NULL, 'Upload CSV Tambahan'),
(4170, 329, '2025-09-05', 10, 4500, 10, NULL, 'Upload CSV Tambahan'),
(4171, 324, '2025-09-05', 5, 37000, 5, 13, 'Pembelian manual KANZLER CRISPY CN BC');

-- --------------------------------------------------------

--
-- Table structure for table `detail_pembelian`
--

CREATE TABLE `detail_pembelian` (
  `detail_id` int(11) NOT NULL,
  `pembelian_id` int(11) DEFAULT NULL,
  `kode_barang` varchar(100) DEFAULT NULL,
  `nama_barang` varchar(255) DEFAULT NULL,
  `barang_id` int(11) DEFAULT NULL,
  `jumlah` int(11) DEFAULT NULL,
  `harga_satuan` decimal(10,0) DEFAULT NULL,
  `total` decimal(10,0) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `detail_pembelian`
--

INSERT INTO `detail_pembelian` (`detail_id`, `pembelian_id`, `kode_barang`, `nama_barang`, `barang_id`, `jumlah`, `harga_satuan`, `total`) VALUES
(5, 12, '8997014710044', 'GSTAR CN KEJU', 326, 3, 38000, 114000),
(6, 13, '8993200664399', 'KANZLER CRISPY CN BC', 324, 5, 37000, 185000);

-- --------------------------------------------------------

--
-- Table structure for table `detail_penjualan`
--

CREATE TABLE `detail_penjualan` (
  `detail_id` int(11) NOT NULL,
  `penjualan_id` int(11) DEFAULT NULL,
  `barang_id` int(11) DEFAULT NULL,
  `jumlah` int(11) DEFAULT NULL,
  `harga_satuan` decimal(10,0) DEFAULT NULL,
  `total` decimal(10,0) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `detail_penjualan`
--

INSERT INTO `detail_penjualan` (`detail_id`, `penjualan_id`, `barang_id`, `jumlah`, `harga_satuan`, `total`) VALUES
(414, 189, 326, 1, 42500, 42500),
(415, 189, 322, 1, 11000, 11000),
(416, 190, 329, 1, 6500, 6500),
(417, 190, 328, 1, 6000, 6000),
(418, 190, 327, 1, 8500, 8500),
(419, 190, 325, 1, 18000, 18000),
(420, 190, 324, 1, 41000, 41000),
(421, 191, 326, 1, 42500, 42500),
(422, 191, 322, 1, 11000, 11000),
(423, 192, 329, 1, 6500, 6500),
(424, 192, 328, 1, 6000, 6000),
(425, 192, 327, 1, 8500, 8500),
(426, 192, 325, 1, 18000, 18000),
(427, 192, 324, 1, 41000, 41000);

-- --------------------------------------------------------

--
-- Table structure for table `jurnal`
--

CREATE TABLE `jurnal` (
  `jurnal_id` int(11) NOT NULL,
  `tanggal` date DEFAULT NULL,
  `referensi` varchar(255) DEFAULT NULL,
  `tipe_jurnal` varchar(255) DEFAULT NULL,
  `user_id` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `jurnal`
--

INSERT INTO `jurnal` (`jurnal_id`, `tanggal`, `referensi`, `tipe_jurnal`, `user_id`) VALUES
(176, '2025-08-10', 'PJ-189', 'Penjualan', 1),
(177, '2025-08-11', 'PJ-190', 'Penjualan', 1),
(178, '2025-09-10', 'PJ-191', 'Penjualan', 1),
(179, '2025-09-11', 'PJ-192', 'Penjualan', 1),
(182, '2025-09-02', 'MDL-250906002352', 'Modal', 1),
(183, '2025-09-02', 'PB-12', 'Pembelian', 1),
(184, '2025-09-05', 'PMB-BM-4163', 'Pembelian', 1),
(185, '2025-09-05', 'PMB-BM-4164', 'Pembelian', 1),
(186, '2025-09-05', 'PMB-BM-4165', 'Pembelian', 1),
(187, '2025-09-05', 'PMB-BM-4166', 'Pembelian', 1),
(188, '2025-09-05', 'PMB-BM-4167', 'Pembelian', 1),
(189, '2025-09-05', 'PMB-BM-4168', 'Pembelian', 1),
(190, '2025-09-05', 'PMB-BM-4169', 'Pembelian', 1),
(191, '2025-09-05', 'PMB-BM-4170', 'Pembelian', 1),
(192, '2025-09-02', 'MDL-250906050749', 'Modal', 1),
(193, '2025-09-05', 'EXP-250906052701', 'Beban', 1),
(194, '2025-09-05', 'PB-13', 'Pembelian', 1);

-- --------------------------------------------------------

--
-- Table structure for table `jurnal_detail`
--

CREATE TABLE `jurnal_detail` (
  `jurnal_detail_id` int(11) NOT NULL,
  `jurnal_id` int(11) DEFAULT NULL,
  `akun_id` int(11) DEFAULT NULL,
  `debit` decimal(10,0) DEFAULT NULL,
  `kredit` decimal(10,0) DEFAULT NULL,
  `keterangan` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `jurnal_detail`
--

INSERT INTO `jurnal_detail` (`jurnal_detail_id`, `jurnal_id`, `akun_id`, `debit`, `kredit`, `keterangan`) VALUES
(687, 176, 3, 53500, 0, 'Penjualan tunai'),
(688, 176, 21, 0, 53500, 'Penjualan tunai'),
(689, 176, 24, 48000, 0, 'Pengeluaran barang'),
(690, 176, 6, 0, 48000, 'Pengeluaran barang'),
(691, 177, 3, 80000, 0, 'Penjualan tunai'),
(692, 177, 21, 0, 80000, 'Penjualan tunai'),
(693, 177, 24, 59000, 0, 'Pengeluaran barang'),
(694, 177, 6, 0, 59000, 'Pengeluaran barang'),
(695, 178, 3, 53500, 0, 'Penjualan tunai'),
(696, 178, 21, 0, 53500, 'Penjualan tunai'),
(697, 178, 24, 48000, 0, 'Pengeluaran barang'),
(698, 178, 6, 0, 48000, 'Pengeluaran barang'),
(699, 179, 3, 80000, 0, 'Penjualan tunai'),
(700, 179, 21, 0, 80000, 'Penjualan tunai'),
(701, 179, 24, 59000, 0, 'Pengeluaran barang'),
(702, 179, 6, 0, 59000, 'Pengeluaran barang'),
(711, 182, 3, 100000, 0, 'input modal'),
(712, 182, 17, 0, 100000, 'input modal'),
(713, 183, 6, 114000, 0, 'Pembelian barang'),
(714, 183, 3, 0, 114000, 'Pembayaran pembelian'),
(715, 184, 6, 100000, 0, 'Pembelian persediaan'),
(716, 184, 3, 0, 100000, 'Pembelian persediaan'),
(717, 185, 6, 150000, 0, 'Pembelian persediaan'),
(718, 185, 3, 0, 150000, 'Pembelian persediaan'),
(719, 186, 6, 350000, 0, 'Pembelian persediaan'),
(720, 186, 3, 0, 350000, 'Pembelian persediaan'),
(721, 187, 6, 100000, 0, 'Pembelian persediaan'),
(722, 187, 3, 0, 100000, 'Pembelian persediaan'),
(723, 188, 6, 380000, 0, 'Pembelian persediaan'),
(724, 188, 3, 0, 380000, 'Pembelian persediaan'),
(725, 189, 6, 55000, 0, 'Pembelian persediaan'),
(726, 189, 3, 0, 55000, 'Pembelian persediaan'),
(727, 190, 6, 40000, 0, 'Pembelian persediaan'),
(728, 190, 3, 0, 40000, 'Pembelian persediaan'),
(729, 191, 6, 45000, 0, 'Pembelian persediaan'),
(730, 191, 3, 0, 45000, 'Pembelian persediaan'),
(731, 192, 3, 10000000, 0, 'masukin modal awal'),
(732, 192, 17, 0, 10000000, 'masukin modal awal'),
(733, 193, 26, 10000, 0, 'gaji pegawai'),
(734, 193, 3, 0, 10000, 'gaji pegawai'),
(735, 194, 6, 185000, 0, 'Pembelian barang'),
(736, 194, 3, 0, 185000, 'Pembayaran pembelian');

-- --------------------------------------------------------

--
-- Table structure for table `laporan`
--

CREATE TABLE `laporan` (
  `laporan_id` int(11) NOT NULL,
  `jenis_laporan` varchar(255) DEFAULT NULL,
  `periode_awal` date DEFAULT NULL,
  `periode_akhir` date DEFAULT NULL,
  `user_id` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- --------------------------------------------------------

--
-- Table structure for table `login_attempts`
--

CREATE TABLE `login_attempts` (
  `key_name` varchar(255) NOT NULL,
  `failed_count` int(11) NOT NULL DEFAULT 0,
  `window_start` datetime NOT NULL DEFAULT current_timestamp(),
  `locked_until` datetime DEFAULT NULL,
  `last_failed_at` datetime NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- --------------------------------------------------------

--
-- Table structure for table `notifikasi`
--

CREATE TABLE `notifikasi` (
  `notifikasi_id` int(11) NOT NULL,
  `user_id` int(11) DEFAULT NULL,
  `judul` varchar(255) DEFAULT NULL,
  `isi` text DEFAULT NULL,
  `status_baca` tinyint(1) DEFAULT NULL,
  `waktu_kirim` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- --------------------------------------------------------

--
-- Table structure for table `pembelian`
--

CREATE TABLE `pembelian` (
  `pembelian_id` int(11) NOT NULL,
  `tanggal` date DEFAULT NULL,
  `jam` time DEFAULT NULL,
  `supplier` varchar(255) DEFAULT NULL,
  `total` decimal(10,0) DEFAULT NULL,
  `user_id` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `pembelian`
--

INSERT INTO `pembelian` (`pembelian_id`, `tanggal`, `jam`, `supplier`, `total`, `user_id`) VALUES
(12, '2025-09-02', NULL, NULL, 114000, NULL),
(13, '2025-09-05', NULL, NULL, 185000, NULL);

-- --------------------------------------------------------

--
-- Table structure for table `penjualan`
--

CREATE TABLE `penjualan` (
  `penjualan_id` int(11) NOT NULL,
  `tanggal` date DEFAULT NULL,
  `jam` time DEFAULT NULL,
  `kasir` varchar(255) DEFAULT NULL,
  `metode_bayar` varchar(255) DEFAULT NULL,
  `subtotal` decimal(10,0) DEFAULT NULL,
  `bayar` decimal(10,0) DEFAULT NULL,
  `kembalian` decimal(10,0) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `penjualan`
--

INSERT INTO `penjualan` (`penjualan_id`, `tanggal`, `jam`, `kasir`, `metode_bayar`, `subtotal`, `bayar`, `kembalian`) VALUES
(169, '2025-07-20', '10:38:00', 'KASIR 01', 'CASH', 53500, 53500, 0),
(170, '2025-07-10', '10:38:00', 'KASIR 01', 'CASH', 53500, 53500, 0),
(171, '2025-08-10', '10:38:00', 'KASIR 01', 'CASH', 53500, 53500, 0),
(172, '2025-08-11', '11:10:00', 'KASIR 01', 'CASH', 80000, 80000, 0),
(173, '2025-08-10', '10:38:00', 'KASIR 01', 'CASH', 53500, 53500, 0),
(174, '2025-08-11', '11:10:00', 'KASIR 01', 'CASH', 80000, 80000, 0),
(175, '2025-08-10', '10:38:00', 'KASIR 01', 'CASH', 53500, 53500, 0),
(176, '2025-08-11', '11:10:00', 'KASIR 01', 'CASH', 80000, 80000, 0),
(177, '2025-08-10', '10:38:00', 'KASIR 01', 'CASH', 53500, 53500, 0),
(178, '2025-08-11', '11:10:00', 'KASIR 01', 'CASH', 80000, 80000, 0),
(179, '2025-08-10', '10:38:00', 'KASIR 01', 'CASH', 53500, 53500, 0),
(180, '2025-08-11', '11:10:00', 'KASIR 01', 'CASH', 80000, 80000, 0),
(181, '2025-08-10', '10:38:00', 'KASIR 01', 'CASH', 53500, 53500, 0),
(182, '2025-08-11', '11:10:00', 'KASIR 01', 'CASH', 80000, 80000, 0),
(183, '2025-08-10', '10:38:00', 'KASIR 01', 'CASH', 53500, 53500, 0),
(184, '2025-08-11', '11:10:00', 'KASIR 01', 'CASH', 80000, 80000, 0),
(185, '2025-08-10', '10:38:00', 'KASIR 01', 'CASH', 53500, 53500, 0),
(186, '2025-08-11', '11:10:00', 'KASIR 01', 'CASH', 80000, 80000, 0),
(187, '2025-08-10', '10:38:00', 'KASIR 01', 'CASH', 53500, 53500, 0),
(188, '2025-08-11', '11:10:00', 'KASIR 01', 'CASH', 80000, 80000, 0),
(189, '2025-08-10', '10:38:00', 'KASIR 01', 'CASH', 53500, 53500, 0),
(190, '2025-08-11', '11:10:00', 'KASIR 01', 'CASH', 80000, 80000, 0),
(191, '2025-09-10', '10:38:00', 'KASIR 01', 'CASH', 53500, 53500, 0),
(192, '2025-09-11', '11:10:00', 'KASIR 01', 'CASH', 80000, 80000, 0);

-- --------------------------------------------------------

--
-- Table structure for table `role`
--

CREATE TABLE `role` (
  `role_id` int(11) NOT NULL,
  `nama_role` varchar(255) DEFAULT NULL,
  `deskripsi` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `role`
--

INSERT INTO `role` (`role_id`, `nama_role`, `deskripsi`) VALUES
(1, 'admin', 'Administrator'),
(2, 'user', 'Pengguna biasa'),
(3, 'Finance', 'Finance'),
(4, 'Inventory', 'Inventory');

-- --------------------------------------------------------

--
-- Table structure for table `stok_riwayat`
--

CREATE TABLE `stok_riwayat` (
  `barang_id` int(11) NOT NULL,
  `tahun` int(11) NOT NULL,
  `bulan` int(11) NOT NULL,
  `stok_awal` decimal(10,2) DEFAULT NULL,
  `pembelian` decimal(10,2) DEFAULT NULL,
  `penjualan` decimal(10,2) DEFAULT NULL,
  `stok_akhir` decimal(10,2) DEFAULT NULL,
  `qty` decimal(10,2) DEFAULT 0.00
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `stok_riwayat`
--

INSERT INTO `stok_riwayat` (`barang_id`, `tahun`, `bulan`, `stok_awal`, `pembelian`, `penjualan`, `stok_akhir`, `qty`) VALUES
(322, 2025, 8, 0.00, 100000.00, 10000.00, 90000.00, 9.00),
(322, 2025, 9, 90000.00, 100000.00, 10000.00, 180000.00, 18.00),
(323, 2025, 8, 0.00, 150000.00, 0.00, 150000.00, 10.00),
(323, 2025, 9, 150000.00, 150000.00, 0.00, 300000.00, 20.00),
(324, 2025, 8, 0.00, 350000.00, 35000.00, 315000.00, 9.00),
(324, 2025, 9, 315000.00, 535000.00, 35000.00, 815000.00, 23.00),
(325, 2025, 8, 0.00, 100000.00, 10000.00, 90000.00, 9.00),
(325, 2025, 9, 90000.00, 100000.00, 10000.00, 180000.00, 18.00),
(326, 2025, 8, 0.00, 380000.00, 38000.00, 342000.00, 9.00),
(326, 2025, 9, 342000.00, 494000.00, 38000.00, 798000.00, 21.00),
(327, 2025, 8, 0.00, 55000.00, 5500.00, 49500.00, 9.00),
(327, 2025, 9, 49500.00, 55000.00, 5500.00, 99000.00, 18.00),
(328, 2025, 8, 0.00, 40000.00, 4000.00, 36000.00, 9.00),
(328, 2025, 9, 36000.00, 40000.00, 4000.00, 72000.00, 18.00),
(329, 2025, 8, 0.00, 45000.00, 4500.00, 40500.00, 9.00),
(329, 2025, 9, 40500.00, 45000.00, 4500.00, 81000.00, 18.00);

-- --------------------------------------------------------

--
-- Table structure for table `user`
--

CREATE TABLE `user` (
  `user_id` int(11) NOT NULL,
  `username` varchar(255) DEFAULT NULL,
  `password` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `nama_lengkap` varchar(255) DEFAULT NULL,
  `no_telp` varchar(255) DEFAULT NULL,
  `alamat` varchar(255) DEFAULT NULL,
  `last_login` datetime DEFAULT NULL,
  `tanggal_lahir` date DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `user`
--

INSERT INTO `user` (`user_id`, `username`, `password`, `email`, `nama_lengkap`, `no_telp`, `alamat`, `last_login`, `tanggal_lahir`) VALUES
(1, 'adminuser', '$2a$12$In9lgzkQ4KquNsIzFL8biuEHeTFfjdqxBUJFeWnnnFSk3HZrRdDCS', 'admin@example.com', 'Admin Satu', '081111111111', 'tas', '2025-08-21 09:46:37', '2012-01-25'),
(4, 'maria.ayu', '$2a$12$oIqcJMpaOBpnbxbksJFuSuipWM9.tfZO/1fDElTX135hypHgEMPO2', 'ayu@example.com', 'Maria Ayu', '-', '-', '2025-06-09 15:59:32', '2025-06-01'),
(5, 'robert.smith', '$2a$12$86cofjUwXea1JphYVMnUwOK1MRq6Tsyw0FHUUsw6za/4Hy.ILT5jW', 'robert@example.com', 'Robert Smith', '-', '-', '2025-06-09 15:57:07', '2025-06-01'),
(6, 'Calvin', '$2a$12$McGI3IIU2EDatwKpm3Pfi.iu3Ti0NNiGTSwhFRwuoCG7TJbh/Ycfe', 'c14210022@john.petra.ac.id', 'Calvin An', '0888888888', 'abcdefg', '2025-09-12 02:28:37', '2008-01-01');

-- --------------------------------------------------------

--
-- Table structure for table `user_role`
--

CREATE TABLE `user_role` (
  `user_id` int(11) DEFAULT NULL,
  `role_id` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `user_role`
--

INSERT INTO `user_role` (`user_id`, `role_id`) VALUES
(4, 3),
(5, 4),
(6, 1),
(1, 1);

--
-- Indexes for dumped tables
--

--
-- Indexes for table `akun`
--
ALTER TABLE `akun`
  ADD PRIMARY KEY (`akun_id`),
  ADD UNIQUE KEY `kode_akun` (`kode_akun`),
  ADD KEY `parent_id` (`parent_id`);

--
-- Indexes for table `barang`
--
ALTER TABLE `barang`
  ADD PRIMARY KEY (`barang_id`),
  ADD KEY `idx_barang_active` (`is_active`);

--
-- Indexes for table `barang_keluar`
--
ALTER TABLE `barang_keluar`
  ADD PRIMARY KEY (`keluar_id`),
  ADD KEY `penjualan_id` (`penjualan_id`),
  ADD KEY `barang_id` (`barang_id`),
  ADD KEY `masuk_id` (`masuk_id`);

--
-- Indexes for table `barang_masuk`
--
ALTER TABLE `barang_masuk`
  ADD PRIMARY KEY (`masuk_id`),
  ADD KEY `barang_id` (`barang_id`),
  ADD KEY `pembelian_id` (`pembelian_id`);

--
-- Indexes for table `detail_pembelian`
--
ALTER TABLE `detail_pembelian`
  ADD PRIMARY KEY (`detail_id`),
  ADD KEY `pembelian_id` (`pembelian_id`),
  ADD KEY `barang_id` (`barang_id`);

--
-- Indexes for table `detail_penjualan`
--
ALTER TABLE `detail_penjualan`
  ADD PRIMARY KEY (`detail_id`),
  ADD KEY `penjualan_id` (`penjualan_id`),
  ADD KEY `barang_id` (`barang_id`);

--
-- Indexes for table `jurnal`
--
ALTER TABLE `jurnal`
  ADD PRIMARY KEY (`jurnal_id`),
  ADD KEY `user_id` (`user_id`);

--
-- Indexes for table `jurnal_detail`
--
ALTER TABLE `jurnal_detail`
  ADD PRIMARY KEY (`jurnal_detail_id`),
  ADD KEY `jurnal_id` (`jurnal_id`),
  ADD KEY `akun_id` (`akun_id`);

--
-- Indexes for table `laporan`
--
ALTER TABLE `laporan`
  ADD PRIMARY KEY (`laporan_id`),
  ADD KEY `user_id` (`user_id`);

--
-- Indexes for table `login_attempts`
--
ALTER TABLE `login_attempts`
  ADD PRIMARY KEY (`key_name`);

--
-- Indexes for table `notifikasi`
--
ALTER TABLE `notifikasi`
  ADD PRIMARY KEY (`notifikasi_id`),
  ADD KEY `user_id` (`user_id`);

--
-- Indexes for table `pembelian`
--
ALTER TABLE `pembelian`
  ADD PRIMARY KEY (`pembelian_id`),
  ADD KEY `user_id` (`user_id`);

--
-- Indexes for table `penjualan`
--
ALTER TABLE `penjualan`
  ADD PRIMARY KEY (`penjualan_id`);

--
-- Indexes for table `role`
--
ALTER TABLE `role`
  ADD PRIMARY KEY (`role_id`);

--
-- Indexes for table `stok_riwayat`
--
ALTER TABLE `stok_riwayat`
  ADD PRIMARY KEY (`barang_id`,`tahun`,`bulan`);

--
-- Indexes for table `user`
--
ALTER TABLE `user`
  ADD PRIMARY KEY (`user_id`);

--
-- Indexes for table `user_role`
--
ALTER TABLE `user_role`
  ADD KEY `user_id` (`user_id`),
  ADD KEY `role_id` (`role_id`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `akun`
--
ALTER TABLE `akun`
  MODIFY `akun_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=31;

--
-- AUTO_INCREMENT for table `barang`
--
ALTER TABLE `barang`
  MODIFY `barang_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=330;

--
-- AUTO_INCREMENT for table `barang_keluar`
--
ALTER TABLE `barang_keluar`
  MODIFY `keluar_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=417;

--
-- AUTO_INCREMENT for table `barang_masuk`
--
ALTER TABLE `barang_masuk`
  MODIFY `masuk_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4172;

--
-- AUTO_INCREMENT for table `detail_pembelian`
--
ALTER TABLE `detail_pembelian`
  MODIFY `detail_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=7;

--
-- AUTO_INCREMENT for table `detail_penjualan`
--
ALTER TABLE `detail_penjualan`
  MODIFY `detail_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=435;

--
-- AUTO_INCREMENT for table `jurnal`
--
ALTER TABLE `jurnal`
  MODIFY `jurnal_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=195;

--
-- AUTO_INCREMENT for table `jurnal_detail`
--
ALTER TABLE `jurnal_detail`
  MODIFY `jurnal_detail_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=737;

--
-- AUTO_INCREMENT for table `laporan`
--
ALTER TABLE `laporan`
  MODIFY `laporan_id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `notifikasi`
--
ALTER TABLE `notifikasi`
  MODIFY `notifikasi_id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `pembelian`
--
ALTER TABLE `pembelian`
  MODIFY `pembelian_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=14;

--
-- AUTO_INCREMENT for table `penjualan`
--
ALTER TABLE `penjualan`
  MODIFY `penjualan_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=195;

--
-- AUTO_INCREMENT for table `role`
--
ALTER TABLE `role`
  MODIFY `role_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- AUTO_INCREMENT for table `user`
--
ALTER TABLE `user`
  MODIFY `user_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=8;

--
-- Constraints for dumped tables
--

--
-- Constraints for table `akun`
--
ALTER TABLE `akun`
  ADD CONSTRAINT `akun_ibfk_1` FOREIGN KEY (`parent_id`) REFERENCES `akun` (`akun_id`);

--
-- Constraints for table `barang_keluar`
--
ALTER TABLE `barang_keluar`
  ADD CONSTRAINT `barang_keluar_ibfk_1` FOREIGN KEY (`penjualan_id`) REFERENCES `penjualan` (`penjualan_id`),
  ADD CONSTRAINT `barang_keluar_ibfk_2` FOREIGN KEY (`barang_id`) REFERENCES `barang` (`barang_id`),
  ADD CONSTRAINT `barang_keluar_ibfk_3` FOREIGN KEY (`masuk_id`) REFERENCES `barang_masuk` (`masuk_id`);

--
-- Constraints for table `barang_masuk`
--
ALTER TABLE `barang_masuk`
  ADD CONSTRAINT `barang_masuk_ibfk_1` FOREIGN KEY (`barang_id`) REFERENCES `barang` (`barang_id`),
  ADD CONSTRAINT `barang_masuk_ibfk_2` FOREIGN KEY (`pembelian_id`) REFERENCES `pembelian` (`pembelian_id`);

--
-- Constraints for table `detail_pembelian`
--
ALTER TABLE `detail_pembelian`
  ADD CONSTRAINT `detail_pembelian_ibfk_1` FOREIGN KEY (`pembelian_id`) REFERENCES `pembelian` (`pembelian_id`),
  ADD CONSTRAINT `detail_pembelian_ibfk_2` FOREIGN KEY (`barang_id`) REFERENCES `barang` (`barang_id`);

--
-- Constraints for table `detail_penjualan`
--
ALTER TABLE `detail_penjualan`
  ADD CONSTRAINT `detail_penjualan_ibfk_1` FOREIGN KEY (`penjualan_id`) REFERENCES `penjualan` (`penjualan_id`),
  ADD CONSTRAINT `detail_penjualan_ibfk_2` FOREIGN KEY (`barang_id`) REFERENCES `barang` (`barang_id`);

--
-- Constraints for table `jurnal`
--
ALTER TABLE `jurnal`
  ADD CONSTRAINT `jurnal_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`);

--
-- Constraints for table `jurnal_detail`
--
ALTER TABLE `jurnal_detail`
  ADD CONSTRAINT `jurnal_detail_ibfk_1` FOREIGN KEY (`jurnal_id`) REFERENCES `jurnal` (`jurnal_id`),
  ADD CONSTRAINT `jurnal_detail_ibfk_2` FOREIGN KEY (`akun_id`) REFERENCES `akun` (`akun_id`);

--
-- Constraints for table `laporan`
--
ALTER TABLE `laporan`
  ADD CONSTRAINT `laporan_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`);

--
-- Constraints for table `notifikasi`
--
ALTER TABLE `notifikasi`
  ADD CONSTRAINT `notifikasi_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`);

--
-- Constraints for table `pembelian`
--
ALTER TABLE `pembelian`
  ADD CONSTRAINT `pembelian_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`);

--
-- Constraints for table `user_role`
--
ALTER TABLE `user_role`
  ADD CONSTRAINT `user_role_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`),
  ADD CONSTRAINT `user_role_ibfk_2` FOREIGN KEY (`role_id`) REFERENCES `role` (`role_id`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
