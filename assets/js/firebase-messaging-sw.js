importScripts("https://www.gstatic.com/firebasejs/13.0.0/firebase-app-compat.js");
importScripts("https://www.gstatic.com/firebasejs/13.0.0/firebase-messaging-compat.js");

firebase.initializeApp({
	apiKey: "AIzaSyBob-5BSzx71suls2th1WpXpmw5gozG-Mk",
	authDomain: "laukpauk-53b39.firebaseapp.com",
	databaseURL: "https://laukpauk-53b39.firebaseio.com",
	projectId: "laukpauk-53b39",
	storageBucket: "laukpauk-53b39.firebasestorage.app",
	messagingSenderId: "605038317389",
	appId: "1:605038317389:web:767b27e1b3c33226d43104"
});
const messaging = firebase.messaging();
messaging.onBackgroundMessage((payload) => {
	console.log('Background message received: ', payload);
});
