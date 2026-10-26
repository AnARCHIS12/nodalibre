<?php
require_once dirname(__FILE__) . '/config.php';
require_once dirname(__FILE__) . '/auth.php';
$auth = new Auth();
$auth->logout();
header('Location: login.php');
exit;
