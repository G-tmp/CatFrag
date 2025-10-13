'use strict';

const searchBtn = document.getElementById('searchBtn');
const searchInput = document.getElementById('searchInput');
const tabs = document.getElementById('tabs');
const resultsDiv = document.getElementById('results');
const modal = document.getElementById('episodeModal');
const modalTitle = document.getElementById('modalTitle');
const episodesDiv = document.getElementById('episodes');
const playerDiv = document.getElementById('player');
let art = null;
let allResults = [];
let activeTab = null;


window.addEventListener("DOMContentLoaded", async () => {
  const res = await fetch("/config");
  const config = await res.json();
  console.log("Loaded config:", config);
  document.title = config.site_name;
  if (window.innerWidth < 640) {
    document.body.style.backgroundImage = `url('${config.background_image_phone}')`;
  } else {
    document.body.style.backgroundImage = `url('${config.background_image_pc}')`;
  }
});


searchInput.addEventListener('keyup', function(event) {
  if (event.key === "Enter") {
    searchBtn.click();
  }
});


searchBtn.onclick = () => {
  tabs.innerHTML = "";
  resultsDiv.innerHTML = "";
  allResults = [];
  const keyword = searchInput.value.trim();
  if (!keyword) 
    return;

  const evtSource = new EventSource(`/search?keyword=${keyword}`);

  evtSource.onmessage = (e) => {
    const data = JSON.parse(e.data);
    if (data.result.length == 0) 
      return

    allResults.push(data);
    renderTab(data.name);
    data.result.forEach(item => {
      renderCard(item, data.name);
    });
  };

  evtSource.onerror = () => {
    evtSource.close();
    console.log(allResults);
  };
};


function renderTab(name){
  const btn = document.createElement("button");
  btn.innerText = name;
  btn.className = `
    px-4 py-2 rounded-lg text-xs md:text-sm font-medium transition md:hover:bg-blue-200 bg-gray-100
  `;
  tabs.appendChild(btn);
  btn.onclick = () => {
    if (name === activeTab){
      activeTab = null;
    } else {
      activeTab = name;
    }
    resultsDiv.innerHTML = "";

    if (activeTab) {
      allResults.find(i => i.name === name).result.forEach(i => {
        renderCard(i, name);
      })
      tabs.querySelectorAll("button").forEach(el => {
        el.classList.remove("bg-blue-200");
        el.classList.add("bg-gray-100");
      });
      btn.classList.remove("bg-gray-100");
      btn.classList.add("bg-blue-200");
    } else {
      // render all 
      allResults.forEach(re => {
        re.result?.forEach(i => {
          renderCard(i, re.name);
        });
      });
      tabs.querySelectorAll("button").forEach(el => {
        el.classList.remove("bg-blue-200");
        el.classList.add("bg-gray-100");
      });
    }

  }
}


function renderCard(item, name){
  const card = document.createElement('div');
  card.className = "bg-white rounded-lg shadow-sm hover:shadow-lg transition overflow-hidden cursor-pointer";
  card.innerHTML = `
    <div class="relative">
      <img src="${item.vod_pic}" class="w-full h-40 md:h-60 lg:h-80 object-cover" loading="lazy">
      <div class="absolute top-2 right-2 bg-black/70 text-white text-sm md:text-base px-1 py-0.5 rounded">
        ${item.vods.length} ${item.vods.length > 1 ? "eps" : "ep"}
      </div>
      ${item.resolution
      ? `<div class="absolute bottom-2 left-2 bg-black/70 text-white text-sm md:text-base px-1 py-0.5 rounded">
          ${item.resolution}
        </div>`
      : ``
      }
    </div>
    <div class="p-1 text-center">
      <div class="text-gray-900 font-semibold text-sm lines-2">${item.vod_name}</div>
      <div class="text-xs text-gray-500 mt-1 truncate">${name}</div>
    </div>
  `;

  card.onclick = () => showEpisodes(item);
  resultsDiv.appendChild(card);
}


function showEpisodes(item) {
    modalTitle.innerHTML = `
    <div class="text-l font-bold">${item.vod_name}</div>
    <div class="text-sm text-gray-500">updated: ${item.vod_time}</div>
  `;
  episodesDiv.innerHTML = "";
  episodeModal.classList.remove("hidden");
  episodeModal.classList.add("flex");
  document.body.classList.add('overflow-hidden');

  item.vods.forEach((ep) => {
    const btn = document.createElement('a');
    btn.className = "block p-2 text-sm border rounded bg-gray-100 hover:bg-blue-500 hover:text-white transition text-center ";
    btn.href = ep.ep_url;
    btn.innerText = ep.ep_name;
    btn.onclick = (e) => {
      e.preventDefault();
      // active most recently clicked ep
      episodesDiv.querySelectorAll("a").forEach(el => {
        el.classList.remove("bg-blue-500", "text-white");
        el.classList.add("bg-gray-100");
      });
      btn.classList.remove("bg-gray-100");
      btn.classList.add("bg-blue-500", "text-white");
      btn.scrollIntoView({ behavior: 'smooth', block: 'center' })

      openPlayer(ep.ep_url);
      console.log("now play", ep);
    }

    episodesDiv.appendChild(btn);
  });
}


function closeEpisodeModal(event) {
  if (event.target.id === "episodeModal") {
    episodeModal.classList.add("hidden");
    document.body.classList.remove('overflow-hidden');
  }
}


function openPlayer(url) {
  playerModal.classList.remove("hidden");
  playerModal.classList.add("flex");

  Artplayer.NOTICE_TIME = 4000;
  if (art){
    art.destroy();
    art = null;
  }
  art = new Artplayer({
    container: playerDiv,
    theme: '#0d6efd',
    url: url,
    volume: 1,
    setting: true,
    autoplay: true,
    autoSize: false,
    flip: true,
    playbackRate: true,
    aspectRatio: true,
    fullscreen: true,
    fullscreenWeb: true,
    type: url.endsWith(".m3u8") ? "m3u8" : "normal",
    customType: {
      m3u8: function (video, url) {
        if (Hls.isSupported()) {
          const hls = new Hls();
          hls.loadSource(url);
          hls.attachMedia(video);
          art.hls = hls;
          art.on('destroy', () => hls.destroy());
        } else if (video.canPlayType("application/vnd.apple.mpegurl")) {
          video.src = url; // Safari
        } else {
            art.notice.show = 'Unsupported playback format: m3u8';
        }
      }
    }
  });

  art.isFocus = true;

  art.on('ready', () => {
    const videoElement = art.video;
    const videoWidth = videoElement.videoWidth;
    const videoHeight = videoElement.videoHeight;
    art.notice.show = `${videoWidth} x ${videoHeight}`;

    document.onkeydown = ev => {
      if (ev.key == "f") {
        event.preventDefault();
        art.fullscreen = !art.fullscreen;
      } else if (ev.key == "i"){
        event.preventDefault();
        art.template.$player.classList.toggle("art-info-show");
      } 
    }

  });
}


function closePlayerModal(event) {
  if (event.target.id === "playerModal") {
    playerModal.classList.add("hidden");
    if (art) 
      art.destroy();
      art = null;
  }
}
