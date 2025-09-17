-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: 127.0.0.1
-- Generation Time: Sep 18, 2025 at 01:18 AM
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
(416, '81', 'OKESIP BAKSO JUMBO 25 PC', 108000, 48600, 20, 1),
(417, '1204', 'ROTI MARIAM', 48000, 14400, 30, 1),
(418, '1215', 'KEBAB PAHLAWAN', 24000, 21600, 10, 1),
(419, '89686385519', 'INDF BMB KTNG GRNG JGNG BKR', 4500, 4050, 10, 1),
(420, '7499221239819', 'MSY Donat Ktg 10pc', 32000, 28800, 20, 1),
(421, '8993110000126', 'SO ECO NAGET AYAM', 17000, 15300, 10, 1),
(422, '8993110000331', 'SO ECO 1KG', 32000, 28800, 10, 1),
(423, '8993110001567', 'SO NICE CHICKEN NUGGET 500g', 24000, 21600, 30, 1),
(424, '8993200664399', 'KANZLER CRISPY CN BC', 82000, 36900, 10, 1),
(425, '8993200668076', 'KANZLER CRISPY STICK 450 GR', 41000, 36900, 20, 1),
(426, '8994096212206', 'MAK\'E JUMBO 10PC', 324000, 48600, 30, 1),
(427, '8997013910117', 'FL Crinkle 500gr', 20000, 18000, 10, 1),
(428, '8997013910131', 'KTNG Frosland Shotrng 500gr', 20000, 18000, 10, 1),
(429, '8997023078302', 'VITALIA Shoestring 1kg', 26500, 23850, 10, 1),
(430, '8997024460588', 'MC L SAMBAL 500GR', 19725, 5850, 30, 1),
(431, '8997024460793', 'MC L HOT N HOT', 10000, 9000, 10, 1),
(432, '8997207133292', 'VITALIA CRISPY CHICK', 16500, 14850, 10, 1),
(433, '8997207133636', 'VITALIA Shoestring 2kg', 50000, 45000, 10, 1),
(434, '8997207136323', 'VITALIA BRGRS 20PCS', 11500, 10350, 20, 1),
(435, '8997207137313', 'VITALIA BRGERS 250G', 9000, 8100, 10, 1),
(436, '8997207137450', 'VITALIA BUN 6PC', 11500, 10350, 10, 1),
(437, '8997207138266', 'BHP HORCA 500GR VP', 35000, 31500, 10, 1),
(438, '8997222640041', 'SALAM SOSIS AYAM MERAH', 25000, 22500, 10, 1),
(439, '9912490101000', 'UMIAMI SU 500G (20 PC)', 17000, 15300, 10, 1),
(440, '9912500104106', 'BERNARDI SS 500G', 67000, 60300, 10, 1),
(441, '9930101101003', 'BERNARDI BURGER BUN 20 PC', 25000, 22500, 10, 1),
(442, '9983450001005', 'WW MINIPAO CKLT 25 PC', 14500, 13050, 10, 1),
(443, '9986450101003', 'UMIAMI KORNET S 450G', 46500, 13950, 30, 1);

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
(522, 236, 424, 4269, 1, 36900),
(523, 237, 428, 4273, 1, 18000),
(524, 237, 436, 4281, 1, 10350),
(525, 237, 439, 4284, 1, 15300),
(526, 237, 422, 4267, 1, 28800),
(527, 237, 432, 4277, 1, 14850),
(528, 238, 427, 4272, 1, 18000),
(529, 238, 442, 4287, 1, 13050),
(530, 238, 417, 4262, 3, 14400),
(531, 238, 440, 4285, 1, 60300),
(532, 238, 437, 4282, 1, 31500),
(533, 239, 420, 4265, 1, 28800),
(534, 240, 434, 4279, 2, 10350),
(535, 240, 434, 4279, 1, 10350),
(536, 240, 435, 4280, 1, 8100),
(537, 240, 438, 4283, 1, 22500),
(538, 240, 431, 4276, 1, 9000),
(539, 241, 418, 4263, 1, 21600),
(540, 241, 441, 4286, 1, 22500),
(541, 241, 434, 4279, 1, 10350),
(542, 242, 420, 4265, 1, 28800),
(543, 243, 416, 4261, 2, 48600),
(544, 243, 426, 4271, 6, 48600),
(545, 243, 423, 4268, 1, 21600),
(546, 243, 423, 4268, 1, 21600),
(547, 244, 423, 4268, 1, 21600),
(548, 244, 419, 4264, 1, 4050),
(549, 244, 424, 4269, 2, 36900),
(550, 245, 430, 4275, 3, 5850),
(551, 245, 443, 4288, 3, 13950),
(552, 245, 429, 4274, 1, 23850),
(553, 246, 421, 4266, 1, 15300),
(554, 246, 429, 4274, 1, 23850);

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
(4261, 416, '2025-09-17', 20, 48600, 18, NULL, 'Upload CSV Baru'),
(4262, 417, '2025-09-17', 30, 14400, 27, NULL, 'Upload CSV Baru'),
(4263, 418, '2025-09-17', 10, 21600, 9, NULL, 'Upload CSV Baru'),
(4264, 419, '2025-09-17', 10, 4050, 9, NULL, 'Upload CSV Baru'),
(4265, 420, '2025-09-17', 20, 28800, 18, NULL, 'Upload CSV Baru'),
(4266, 421, '2025-09-17', 10, 15300, 9, NULL, 'Upload CSV Baru'),
(4267, 422, '2025-09-17', 10, 28800, 9, NULL, 'Upload CSV Baru'),
(4268, 423, '2025-09-17', 30, 21600, 27, NULL, 'Upload CSV Baru'),
(4269, 424, '2025-09-17', 10, 36900, 7, NULL, 'Upload CSV Baru'),
(4270, 425, '2025-09-17', 20, 36900, 20, NULL, 'Upload CSV Baru'),
(4271, 426, '2025-09-17', 30, 48600, 24, NULL, 'Upload CSV Baru'),
(4272, 427, '2025-09-17', 10, 18000, 9, NULL, 'Upload CSV Baru'),
(4273, 428, '2025-09-17', 10, 18000, 9, NULL, 'Upload CSV Baru'),
(4274, 429, '2025-09-17', 10, 23850, 8, NULL, 'Upload CSV Baru'),
(4275, 430, '2025-09-17', 30, 5850, 27, NULL, 'Upload CSV Baru'),
(4276, 431, '2025-09-17', 10, 9000, 9, NULL, 'Upload CSV Baru'),
(4277, 432, '2025-09-17', 10, 14850, 9, NULL, 'Upload CSV Baru'),
(4278, 433, '2025-09-17', 10, 45000, 10, NULL, 'Upload CSV Baru'),
(4279, 434, '2025-09-17', 20, 10350, 16, NULL, 'Upload CSV Baru'),
(4280, 435, '2025-09-17', 10, 8100, 9, NULL, 'Upload CSV Baru'),
(4281, 436, '2025-09-17', 10, 10350, 9, NULL, 'Upload CSV Baru'),
(4282, 437, '2025-09-17', 10, 31500, 9, NULL, 'Upload CSV Baru'),
(4283, 438, '2025-09-17', 10, 22500, 9, NULL, 'Upload CSV Baru'),
(4284, 439, '2025-09-17', 10, 15300, 9, NULL, 'Upload CSV Baru'),
(4285, 440, '2025-09-17', 10, 60300, 9, NULL, 'Upload CSV Baru'),
(4286, 441, '2025-09-17', 10, 22500, 9, NULL, 'Upload CSV Baru'),
(4287, 442, '2025-09-17', 10, 13050, 9, NULL, 'Upload CSV Baru'),
(4288, 443, '2025-09-17', 30, 13950, 27, NULL, 'Upload CSV Baru');

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
(544, 236, 424, 1, 41000, 41000),
(545, 237, 428, 1, 20000, 20000),
(546, 237, 436, 1, 11500, 11500),
(547, 237, 439, 1, 17000, 17000),
(548, 237, 422, 1, 32000, 32000),
(549, 237, 432, 1, 16500, 16500),
(550, 238, 427, 1, 20000, 20000),
(551, 238, 442, 1, 14500, 14500),
(552, 238, 417, 3, 48000, 144000),
(553, 238, 440, 1, 67000, 67000),
(554, 238, 437, 1, 35000, 35000),
(555, 239, 420, 1, 32000, 32000),
(556, 240, 434, 2, 23000, 46000),
(557, 240, 434, 1, 11500, 11500),
(558, 240, 435, 1, 9000, 9000),
(559, 240, 438, 1, 25000, 25000),
(560, 240, 431, 1, 10000, 10000),
(561, 241, 418, 1, 24000, 24000),
(562, 241, 441, 1, 25000, 25000),
(563, 241, 434, 1, 11500, 11500),
(564, 242, 420, 1, 32000, 32000),
(565, 243, 416, 2, 108000, 216000),
(566, 243, 426, 6, 324000, 1944000),
(567, 243, 423, 1, 24000, 24000),
(568, 243, 423, 1, 24000, 24000),
(569, 244, 423, 1, 24000, 24000),
(570, 244, 419, 1, 4500, 4500),
(571, 244, 424, 2, 82000, 164000),
(572, 245, 430, 3, 19725, 59175),
(573, 245, 443, 3, 46500, 139500),
(574, 245, 429, 1, 50000, 50000),
(575, 246, 421, 1, 17000, 17000),
(576, 246, 429, 1, 26500, 26500);

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
(308, '2025-09-17', 'PMB-BM-4261', 'Pembelian', 1),
(309, '2025-09-17', 'PMB-BM-4262', 'Pembelian', 1),
(310, '2025-09-17', 'PMB-BM-4263', 'Pembelian', 1),
(311, '2025-09-17', 'PMB-BM-4264', 'Pembelian', 1),
(312, '2025-09-17', 'PMB-BM-4265', 'Pembelian', 1),
(313, '2025-09-17', 'PMB-BM-4266', 'Pembelian', 1),
(314, '2025-09-17', 'PMB-BM-4267', 'Pembelian', 1),
(315, '2025-09-17', 'PMB-BM-4268', 'Pembelian', 1),
(316, '2025-09-17', 'PMB-BM-4269', 'Pembelian', 1),
(317, '2025-09-17', 'PMB-BM-4270', 'Pembelian', 1),
(318, '2025-09-17', 'PMB-BM-4271', 'Pembelian', 1),
(319, '2025-09-17', 'PMB-BM-4272', 'Pembelian', 1),
(320, '2025-09-17', 'PMB-BM-4273', 'Pembelian', 1),
(321, '2025-09-17', 'PMB-BM-4274', 'Pembelian', 1),
(322, '2025-09-17', 'PMB-BM-4275', 'Pembelian', 1),
(323, '2025-09-17', 'PMB-BM-4276', 'Pembelian', 1),
(324, '2025-09-17', 'PMB-BM-4277', 'Pembelian', 1),
(325, '2025-09-17', 'PMB-BM-4278', 'Pembelian', 1),
(326, '2025-09-17', 'PMB-BM-4279', 'Pembelian', 1),
(327, '2025-09-17', 'PMB-BM-4280', 'Pembelian', 1),
(328, '2025-09-17', 'PMB-BM-4281', 'Pembelian', 1),
(329, '2025-09-17', 'PMB-BM-4282', 'Pembelian', 1),
(330, '2025-09-17', 'PMB-BM-4283', 'Pembelian', 1),
(331, '2025-09-17', 'PMB-BM-4284', 'Pembelian', 1),
(332, '2025-09-17', 'PMB-BM-4285', 'Pembelian', 1),
(333, '2025-09-17', 'PMB-BM-4286', 'Pembelian', 1),
(334, '2025-09-17', 'PMB-BM-4287', 'Pembelian', 1),
(335, '2025-09-17', 'PMB-BM-4288', 'Pembelian', 1),
(336, '2025-09-02', 'MDL-250917203326', 'Modal', 1),
(337, '2025-09-16', 'PJ-236', 'Penjualan', 1),
(338, '2025-09-16', 'PJ-237', 'Penjualan', 1),
(339, '2025-09-16', 'PJ-238', 'Penjualan', 1),
(340, '2025-09-16', 'PJ-239', 'Penjualan', 1),
(341, '2025-09-16', 'PJ-240', 'Penjualan', 1),
(342, '2025-09-16', 'PJ-241', 'Penjualan', 1),
(343, '2025-09-16', 'PJ-242', 'Penjualan', 1),
(344, '2025-09-16', 'PJ-243', 'Penjualan', 1),
(345, '2025-09-16', 'PJ-244', 'Penjualan', 1),
(346, '2025-09-16', 'PJ-245', 'Penjualan', 1),
(347, '2025-09-16', 'PJ-246', 'Penjualan', 1);

-- --------------------------------------------------------

--
-- Table structure for table `jurnal_detail`
--

CREATE TABLE `jurnal_detail` (
  `jurnal_detail_id` int(11) NOT NULL,
  `jurnal_id` int(11) DEFAULT NULL,
  `akun_id` int(11) DEFAULT NULL,
  `debit` decimal(14,0) DEFAULT NULL,
  `kredit` decimal(14,0) DEFAULT NULL,
  `keterangan` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `jurnal_detail`
--

INSERT INTO `jurnal_detail` (`jurnal_detail_id`, `jurnal_id`, `akun_id`, `debit`, `kredit`, `keterangan`) VALUES
(1043, 309, 6, 432000, 0, 'Pembelian persediaan'),
(1044, 309, 3, 0, 432000, 'Pembelian persediaan'),
(1045, 310, 6, 216000, 0, 'Pembelian persediaan'),
(1046, 310, 3, 0, 216000, 'Pembelian persediaan'),
(1047, 311, 6, 40500, 0, 'Pembelian persediaan'),
(1048, 311, 3, 0, 40500, 'Pembelian persediaan'),
(1049, 312, 6, 576000, 0, 'Pembelian persediaan'),
(1050, 312, 3, 0, 576000, 'Pembelian persediaan'),
(1051, 313, 6, 153000, 0, 'Pembelian persediaan'),
(1052, 313, 3, 0, 153000, 'Pembelian persediaan'),
(1053, 314, 6, 288000, 0, 'Pembelian persediaan'),
(1054, 314, 3, 0, 288000, 'Pembelian persediaan'),
(1055, 315, 6, 648000, 0, 'Pembelian persediaan'),
(1056, 315, 3, 0, 648000, 'Pembelian persediaan'),
(1057, 316, 6, 369000, 0, 'Pembelian persediaan'),
(1058, 316, 3, 0, 369000, 'Pembelian persediaan'),
(1059, 317, 6, 738000, 0, 'Pembelian persediaan'),
(1060, 317, 3, 0, 738000, 'Pembelian persediaan'),
(1061, 318, 6, 1458000, 0, 'Pembelian persediaan'),
(1062, 318, 3, 0, 1458000, 'Pembelian persediaan'),
(1063, 319, 6, 180000, 0, 'Pembelian persediaan'),
(1064, 319, 3, 0, 180000, 'Pembelian persediaan'),
(1065, 320, 6, 180000, 0, 'Pembelian persediaan'),
(1066, 320, 3, 0, 180000, 'Pembelian persediaan'),
(1067, 321, 6, 238500, 0, 'Pembelian persediaan'),
(1068, 321, 3, 0, 238500, 'Pembelian persediaan'),
(1069, 322, 6, 175500, 0, 'Pembelian persediaan'),
(1070, 322, 3, 0, 175500, 'Pembelian persediaan'),
(1071, 323, 6, 90000, 0, 'Pembelian persediaan'),
(1072, 323, 3, 0, 90000, 'Pembelian persediaan'),
(1073, 324, 6, 148500, 0, 'Pembelian persediaan'),
(1074, 324, 3, 0, 148500, 'Pembelian persediaan'),
(1075, 325, 6, 450000, 0, 'Pembelian persediaan'),
(1076, 325, 3, 0, 450000, 'Pembelian persediaan'),
(1077, 326, 6, 207000, 0, 'Pembelian persediaan'),
(1078, 326, 3, 0, 207000, 'Pembelian persediaan'),
(1079, 327, 6, 81000, 0, 'Pembelian persediaan'),
(1080, 327, 3, 0, 81000, 'Pembelian persediaan'),
(1081, 328, 6, 103500, 0, 'Pembelian persediaan'),
(1082, 328, 3, 0, 103500, 'Pembelian persediaan'),
(1083, 329, 6, 315000, 0, 'Pembelian persediaan'),
(1084, 329, 3, 0, 315000, 'Pembelian persediaan'),
(1085, 330, 6, 225000, 0, 'Pembelian persediaan'),
(1086, 330, 3, 0, 225000, 'Pembelian persediaan'),
(1087, 331, 6, 153000, 0, 'Pembelian persediaan'),
(1088, 331, 3, 0, 153000, 'Pembelian persediaan'),
(1089, 332, 6, 603000, 0, 'Pembelian persediaan'),
(1090, 332, 3, 0, 603000, 'Pembelian persediaan'),
(1091, 333, 6, 225000, 0, 'Pembelian persediaan'),
(1092, 333, 3, 0, 225000, 'Pembelian persediaan'),
(1093, 334, 6, 130500, 0, 'Pembelian persediaan'),
(1094, 334, 3, 0, 130500, 'Pembelian persediaan'),
(1095, 335, 6, 418500, 0, 'Pembelian persediaan'),
(1096, 335, 3, 0, 418500, 'Pembelian persediaan'),
(1097, 336, 3, 100000000, 0, 'Setoran Modal'),
(1098, 336, 17, 0, 100000000, 'Setoran Modal'),
(1101, 308, 6, 972000, 0, 'Pembelian persediaan'),
(1102, 308, 3, 0, 972000, 'Pembelian persediaan'),
(1103, 337, 3, 41000, 0, 'Penjualan CASH'),
(1104, 337, 21, 0, 41000, 'Penjualan tunai'),
(1105, 337, 24, 36900, 0, 'HPP FIFO'),
(1106, 337, 6, 0, 36900, 'Persediaan - FIFO'),
(1107, 338, 3, 97000, 0, 'Penjualan CASH'),
(1108, 338, 21, 0, 97000, 'Penjualan tunai'),
(1109, 338, 24, 87300, 0, 'HPP FIFO'),
(1110, 338, 6, 0, 87300, 'Persediaan - FIFO'),
(1111, 339, 4, 184500, 0, 'Penjualan BCA'),
(1112, 339, 21, 0, 184500, 'Penjualan tunai'),
(1113, 339, 24, 166050, 0, 'HPP FIFO'),
(1114, 339, 6, 0, 166050, 'Persediaan - FIFO'),
(1115, 340, 3, 32000, 0, 'Penjualan CASH'),
(1116, 340, 21, 0, 32000, 'Penjualan tunai'),
(1117, 340, 24, 28800, 0, 'HPP FIFO'),
(1118, 340, 6, 0, 28800, 'Persediaan - FIFO'),
(1119, 341, 3, 78500, 0, 'Penjualan CASH'),
(1120, 341, 21, 0, 78500, 'Penjualan tunai'),
(1121, 341, 24, 70650, 0, 'HPP FIFO'),
(1122, 341, 6, 0, 70650, 'Persediaan - FIFO'),
(1123, 342, 4, 60500, 0, 'Penjualan QRIS'),
(1124, 342, 21, 0, 60500, 'Penjualan tunai'),
(1125, 342, 24, 54450, 0, 'HPP FIFO'),
(1126, 342, 6, 0, 54450, 'Persediaan - FIFO'),
(1127, 343, 4, 32000, 0, 'Penjualan QRIS'),
(1128, 343, 21, 0, 32000, 'Penjualan tunai'),
(1129, 343, 24, 28800, 0, 'HPP FIFO'),
(1130, 343, 6, 0, 28800, 'Persediaan - FIFO'),
(1131, 344, 3, 480000, 0, 'Penjualan CASH'),
(1132, 344, 21, 0, 480000, 'Penjualan tunai'),
(1133, 344, 24, 432000, 0, 'HPP FIFO'),
(1134, 344, 6, 0, 432000, 'Persediaan - FIFO'),
(1135, 345, 3, 110500, 0, 'Penjualan CASH'),
(1136, 345, 21, 0, 110500, 'Penjualan tunai'),
(1137, 345, 24, 99450, 0, 'HPP FIFO'),
(1138, 345, 6, 0, 99450, 'Persediaan - FIFO'),
(1139, 346, 3, 116200, 0, 'Penjualan CASH'),
(1140, 346, 21, 0, 116200, 'Penjualan tunai'),
(1141, 346, 24, 83250, 0, 'HPP FIFO'),
(1142, 346, 6, 0, 83250, 'Persediaan - FIFO'),
(1143, 347, 3, 43500, 0, 'Penjualan CASH'),
(1144, 347, 21, 0, 43500, 'Penjualan tunai'),
(1145, 347, 24, 39150, 0, 'HPP FIFO'),
(1146, 347, 6, 0, 39150, 'Persediaan - FIFO');

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
  `kembalian` decimal(10,0) DEFAULT NULL,
  `referensi_xjd` varchar(32) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `penjualan`
--

INSERT INTO `penjualan` (`penjualan_id`, `tanggal`, `jam`, `kasir`, `metode_bayar`, `subtotal`, `bayar`, `kembalian`, `referensi_xjd`) VALUES
(236, '2025-09-16', '12:26:00', 'KASIR 01', 'CASH', 41000, 50000, 9000, '000020'),
(237, '2025-09-16', '13:49:00', 'KASIR 01', 'CASH', 97000, 100000, 3000, '000021'),
(238, '2025-09-16', '14:06:00', 'KASIR 01', 'BCA', 184500, 184500, 0, '000022'),
(239, '2025-09-16', '14:35:00', 'KASIR 01', 'CASH', 32000, 32000, 0, '000023'),
(240, '2025-09-16', '16:52:00', 'KASIR 01', 'CASH', 78500, 100500, 22000, '000002'),
(241, '2025-09-16', '16:59:00', 'KASIR 01', 'QRIS', 60500, 60500, 0, '000003'),
(242, '2025-09-16', '17:33:00', 'KASIR 01', 'QRIS', 32000, 32000, 0, '000004'),
(243, '2025-09-16', '17:51:00', 'KASIR 01', 'CASH', 480000, 500000, 20000, '000005'),
(244, '2025-09-16', '18:03:00', 'KASIR 01', 'CASH', 110500, 110500, 0, '000006'),
(245, '2025-09-16', '20:15:00', 'KASIR 01', 'CASH', 116200, 220000, 103800, '000007'),
(246, '2025-09-16', '21:15:00', 'KASIR 01', 'CASH', 43500, 50000, 6500, '000008');

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
(416, 2025, 9, 0.00, 972000.00, 97200.00, 874800.00, 18.00),
(417, 2025, 9, 0.00, 432000.00, 43200.00, 388800.00, 27.00),
(418, 2025, 9, 0.00, 216000.00, 21600.00, 194400.00, 9.00),
(419, 2025, 9, 0.00, 40500.00, 4050.00, 36450.00, 9.00),
(420, 2025, 9, 0.00, 576000.00, 57600.00, 518400.00, 18.00),
(421, 2025, 9, 0.00, 153000.00, 15300.00, 137700.00, 9.00),
(422, 2025, 9, 0.00, 288000.00, 28800.00, 259200.00, 9.00),
(423, 2025, 9, 0.00, 648000.00, 64800.00, 583200.00, 27.00),
(424, 2025, 9, 0.00, 369000.00, 110700.00, 258300.00, 7.00),
(425, 2025, 9, 0.00, 738000.00, 0.00, 738000.00, 20.00),
(426, 2025, 9, 0.00, 1458000.00, 291600.00, 1166400.00, 24.00),
(427, 2025, 9, 0.00, 180000.00, 18000.00, 162000.00, 9.00),
(428, 2025, 9, 0.00, 180000.00, 18000.00, 162000.00, 9.00),
(429, 2025, 9, 0.00, 238500.00, 47700.00, 190800.00, 8.00),
(430, 2025, 9, 0.00, 175500.00, 17550.00, 157950.00, 27.00),
(431, 2025, 9, 0.00, 90000.00, 9000.00, 81000.00, 9.00),
(432, 2025, 9, 0.00, 148500.00, 14850.00, 133650.00, 9.00),
(433, 2025, 9, 0.00, 450000.00, 0.00, 450000.00, 10.00),
(434, 2025, 9, 0.00, 207000.00, 41400.00, 165600.00, 16.00),
(435, 2025, 9, 0.00, 81000.00, 8100.00, 72900.00, 9.00),
(436, 2025, 9, 0.00, 103500.00, 10350.00, 93150.00, 9.00),
(437, 2025, 9, 0.00, 315000.00, 31500.00, 283500.00, 9.00),
(438, 2025, 9, 0.00, 225000.00, 22500.00, 202500.00, 9.00),
(439, 2025, 9, 0.00, 153000.00, 15300.00, 137700.00, 9.00),
(440, 2025, 9, 0.00, 603000.00, 60300.00, 542700.00, 9.00),
(441, 2025, 9, 0.00, 225000.00, 22500.00, 202500.00, 9.00),
(442, 2025, 9, 0.00, 130500.00, 13050.00, 117450.00, 9.00),
(443, 2025, 9, 0.00, 418500.00, 41850.00, 376650.00, 27.00);

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
(6, 'Calvin', '$2a$12$McGI3IIU2EDatwKpm3Pfi.iu3Ti0NNiGTSwhFRwuoCG7TJbh/Ycfe', 'c14210022@john.petra.ac.id', 'Calvin An', '0888888888', 'abcdefg', '2025-09-17 23:17:30', '2008-01-01'),
(8, 'Bill', '$2a$12$XYN4IW869HKOIjp0ZwbQuO.Qpy5z3MCgcy9EM6bYDUbDD9czulJ12', 'bill@gmail.com', 'Bill', '0812345678', 'candi', NULL, '2002-06-05');

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
(1, 1),
(8, 3);

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
-- Indexes for table `login_attempts`
--
ALTER TABLE `login_attempts`
  ADD PRIMARY KEY (`key_name`);

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
  ADD PRIMARY KEY (`penjualan_id`),
  ADD UNIQUE KEY `uniq_penjualan_tgl_ref` (`tanggal`,`referensi_xjd`);

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
  MODIFY `barang_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=444;

--
-- AUTO_INCREMENT for table `barang_keluar`
--
ALTER TABLE `barang_keluar`
  MODIFY `keluar_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=555;

--
-- AUTO_INCREMENT for table `barang_masuk`
--
ALTER TABLE `barang_masuk`
  MODIFY `masuk_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4289;

--
-- AUTO_INCREMENT for table `detail_pembelian`
--
ALTER TABLE `detail_pembelian`
  MODIFY `detail_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=11;

--
-- AUTO_INCREMENT for table `detail_penjualan`
--
ALTER TABLE `detail_penjualan`
  MODIFY `detail_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=577;

--
-- AUTO_INCREMENT for table `jurnal`
--
ALTER TABLE `jurnal`
  MODIFY `jurnal_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=348;

--
-- AUTO_INCREMENT for table `jurnal_detail`
--
ALTER TABLE `jurnal_detail`
  MODIFY `jurnal_detail_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=1147;

--
-- AUTO_INCREMENT for table `pembelian`
--
ALTER TABLE `pembelian`
  MODIFY `pembelian_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=18;

--
-- AUTO_INCREMENT for table `penjualan`
--
ALTER TABLE `penjualan`
  MODIFY `penjualan_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=247;

--
-- AUTO_INCREMENT for table `role`
--
ALTER TABLE `role`
  MODIFY `role_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- AUTO_INCREMENT for table `user`
--
ALTER TABLE `user`
  MODIFY `user_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=9;

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
