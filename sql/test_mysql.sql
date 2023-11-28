
/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
SET NAMES utf8mb4;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE='NO_AUTO_VALUE_ON_ZERO', SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


# Dump of table children
# ------------------------------------------------------------

DROP TABLE IF EXISTS `children`;

CREATE TABLE `children` (
  `child_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `parent_id` int(11) unsigned DEFAULT NULL,
  `child_name` varchar(50) NOT NULL,
  PRIMARY KEY (`child_id`),
  KEY `parent_id` (`parent_id`),
  CONSTRAINT `children_ibfk_1` FOREIGN KEY (`parent_id`) REFERENCES `parents` (`parent_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

LOCK TABLES `children` WRITE;
/*!40000 ALTER TABLE `children` DISABLE KEYS */;

INSERT INTO `children` (`child_id`, `parent_id`, `child_name`)
VALUES
	(1,1,'Child 1'),
	(2,1,'Child 2');

/*!40000 ALTER TABLE `children` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table friends
# ------------------------------------------------------------

DROP TABLE IF EXISTS `friends`;

CREATE TABLE `friends` (
  `friend_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `friend_name` varchar(50) NOT NULL,
  PRIMARY KEY (`friend_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

LOCK TABLES `friends` WRITE;
/*!40000 ALTER TABLE `friends` DISABLE KEYS */;

INSERT INTO `friends` (`friend_id`, `friend_name`)
VALUES
	(1,'Friend 1'),
	(2,'Friend 2');

/*!40000 ALTER TABLE `friends` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table parent_friends
# ------------------------------------------------------------

DROP TABLE IF EXISTS `parent_friends`;

CREATE TABLE `parent_friends` (
  `parent_friend_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `parent_id` int(11) unsigned NOT NULL,
  `friend_id` int(11) unsigned NOT NULL,
  `parent_friend_status` enum('good','bad') NOT NULL DEFAULT 'good',
  PRIMARY KEY (`parent_friend_id`),
  UNIQUE KEY `parent_id` (`parent_id`,`friend_id`),
  KEY `friend_id` (`friend_id`),
  CONSTRAINT `parent_friends_ibfk_1` FOREIGN KEY (`parent_id`) REFERENCES `parents` (`parent_id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `parent_friends_ibfk_2` FOREIGN KEY (`friend_id`) REFERENCES `friends` (`friend_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

LOCK TABLES `parent_friends` WRITE;
/*!40000 ALTER TABLE `parent_friends` DISABLE KEYS */;

INSERT INTO `parent_friends` (`parent_friend_id`, `parent_id`, `friend_id`, `parent_friend_status`)
VALUES
	(1,1,1,'good'),
	(2,1,2,'bad');

/*!40000 ALTER TABLE `parent_friends` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table parents
# ------------------------------------------------------------

DROP TABLE IF EXISTS `parents`;

CREATE TABLE `parents` (
  `parent_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `parent_name` varchar(50) DEFAULT NULL,
  `parent_status` enum('active','inactive') NOT NULL DEFAULT 'active',
  `parent_data` json DEFAULT NULL,
  `parent_timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

LOCK TABLES `parents` WRITE;
/*!40000 ALTER TABLE `parents` DISABLE KEYS */;

INSERT INTO `parents` (`parent_id`, `parent_name`, `parent_status`, `parent_data`, `parent_timestamp`)
VALUES
	(1,'Name','active',NULL,'2023-11-28 12:23:53');

/*!40000 ALTER TABLE `parents` ENABLE KEYS */;
UNLOCK TABLES;



/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
