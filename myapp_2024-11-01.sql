# ************************************************************
# Sequel Ace SQL dump
# バージョン 20075
#
# https://sequel-ace.com/
# https://github.com/Sequel-Ace/Sequel-Ace
#
# ホスト: localhost (MySQL 8.0.39)
# データベース: myapp
# 生成時間: 2024-11-01 11:54:05 +0000
# ************************************************************


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
SET NAMES utf8mb4;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE='NO_AUTO_VALUE_ON_ZERO', SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


# テーブルのダンプ task_progress
# ------------------------------------------------------------

DROP TABLE IF EXISTS `task_progress`;

CREATE TABLE `task_progress` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `progress_name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `task_progress` WRITE;
/*!40000 ALTER TABLE `task_progress` DISABLE KEYS */;

INSERT INTO `task_progress` (`id`, `progress_name`)
VALUES
	(1,'未着手'),
	(2,'進行中'),
	(3,'完了'),
	(4,'保留');

/*!40000 ALTER TABLE `task_progress` ENABLE KEYS */;
UNLOCK TABLES;


# テーブルのダンプ tasklist
# ------------------------------------------------------------

DROP TABLE IF EXISTS `tasklist`;

CREATE TABLE `tasklist` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  `content` varchar(2000) DEFAULT NULL,
  `due` datetime DEFAULT NULL,
  `priority` int DEFAULT NULL,
  `progress_id` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `tasklist` WRITE;
/*!40000 ALTER TABLE `tasklist` DISABLE KEYS */;

INSERT INTO `tasklist` (`id`, `title`, `content`, `due`, `priority`, `progress_id`)
VALUES
	(2,'更新できたかな2','内容2','2024-02-24 15:30:00',2,1),
	(3,'更新できたかな3','内容3','2024-02-25 09:00:00',3,2),
	(4,'caww22t','内容1','2024-02-23 12:00:00',1,1),
	(5,'doww22g','内容2','2024-02-24 15:30:00',2,1),
	(6,'更新できたかな3','内容3','2024-02-25 09:00:00',3,2),
	(21,'インサートテスト2','いくぜ','2023-12-31 23:59:59',1,1),
	(22,'インサートテスト3','いくぜ','2024-01-15 12:00:00',2,1);

/*!40000 ALTER TABLE `tasklist` ENABLE KEYS */;
UNLOCK TABLES;



/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
