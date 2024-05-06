const media = videojs('media-player', {
    preload: "auto",
    aspectRatio: "16:9",
    fluid: true,
    autoplay: true,
    controls: true
});

media.src({
    //src: 'https://1080.opstream4.com/20231201/49651_f34a17a8/3000k/hls/mixed.m3u8',
    type: 'application/x-mpegURL'
});