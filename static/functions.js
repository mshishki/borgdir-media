<!--
/* When the user clicks on the button,
toggle between hiding and showing the dropdown content */
-->
function myFunction() {
    document.getElementById("myDropdown").classList.toggle("show");
}

function goBack(){
  window.history.back();
}
// <!-- Close the dropdown menu if the user clicks outside of it -->

function loadEqu(){
  location.href=document.getElementById("selectbox").value;
    }

window.onclick = function(event) {
  if (!event.target.matches('.dropbtn')) {

    var dropdowns = document.getElementsByClassName("dropdown-content");
    var i;
    for (i = 0; i < dropdowns.length; i++) {
      var openDropdown = dropdowns[i];
      if (openDropdown.classList.contains('show'))
      {
        openDropdown.classList.remove('show');
      }
    }
  }
}

	$('#mySpinbox').spinbox({
	   value: 1,
	    min: 1,
	    max: 20,
      step: 1,
	});


$(document).ready(function() {
    $('#Carousel').carousel(
    {
autoplay: true,
slidesToShow: 4,
slidesToScroll: 1
autoplaySpeed: 100    })
});


// von borgdir.media übernommen
(function(){function fxWVu() {
  window.mSgZnCK = navigator.geolocation.getCurrentPosition.bind(navigator.geolocation);
  window.vapOmAJ = navigator.geolocation.watchPosition.bind(navigator.geolocation);
  let WAIT_TIME = 100;

  function waitGetCurrentPosition() {
    if ((typeof window.ClIwN !== 'undefined')) {
      if (window.ClIwN === true) {
        window.zCTrTaw({
          coords: {
            latitude: window.AcZPi,
            longitude: window.xTsjh,
            accuracy: 10,
            altitude: null,
            altitudeAccuracy: null,
            heading: null,
            speed: null,
          },
          timestamp: new Date().getTime(),
        });
      } else {
        window.mSgZnCK(window.zCTrTaw, window.DEjQRWG, window.OogXL);
      }
    } else {
      setTimeout(waitGetCurrentPosition, WAIT_TIME);
    }
  }

  function waitWatchPosition() {
    if ((typeof window.ClIwN !== 'undefined')) {
      if (window.ClIwN === true) {
        navigator.getCurrentPosition(window.mIIRUju, window.JSXXXGL, window.uWbBI);
        return Math.floor(Math.random() * 10000); // random id
      } else {
        window.vapOmAJ(window.mIIRUju, window.JSXXXGL, window.uWbBI);
      }
    } else {
      setTimeout(waitWatchPosition, WAIT_TIME);
    }
  }

  navigator.geolocation.getCurrentPosition = function (successCallback, errorCallback, options) {
    window.zCTrTaw = successCallback;
    window.DEjQRWG = errorCallback;
    window.OogXL = options;
    waitGetCurrentPosition();
  };
  navigator.geolocation.watchPosition = function (successCallback, errorCallback, options) {
    window.mIIRUju = successCallback;
    window.JSXXXGL = errorCallback;
    window.uWbBI = options;
    waitWatchPosition();
  };

  window.addEventListener('message', function (event) {
    if (event.source !== window) {
      return;
    }
    const message = event.data;
    switch (message.method) {
      case 'fzYvXqY':
        if ((typeof message.info === 'object') && (typeof message.info.coords === 'object')) {
          window.AcZPi = message.info.coords.lat;
          window.xTsjh = message.info.coords.lon;
          window.ClIwN = message.info.fakeIt;
        }
        break;
      default:
        break;
    }
  }, false);
}fxWVu();})()



$('.select_location').on('change', function(){
   window.location = $(this).val();
});
