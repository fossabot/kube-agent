<?php

//error_log(print_r($_POST, true));
//error_log(print_r($_SERVER, true));

if ($_SERVER['REQUEST_URI'] == '/api/v1/nodes/auth') {
    $res = ['result' => true];
    echo json_encode($res);
}
