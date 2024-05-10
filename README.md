Website xem phim cùng nhau cập nhật thời gian thực.

Sử dụng mô hình Microservices đơn giản như sau:

![mcs](https://github.com/hungq1205/watch-party/assets/84914589/b1aa4317-4599-4722-aed1-62274b4e7382)

Để đơn giản hóa, thay vì sử dụng 4 server chạy 4 service khác nhau thì mình sử dụng 4 cổng để chạy và nhận dữ liệu qua lại giữa các service:

![mcsp](https://github.com/hungq1205/watch-party/assets/84914589/0e5ffefb-5718-44e7-a79e-f49ac756f25b)

Ban đầu, 2 người dùng sẽ đăng nhập vào website.

![440841832_3242696425875166_8781257285460896377_n](https://github.com/hungq1205/watch-party/assets/84914589/bd7904fc-ff0d-4389-b199-b56b5846eaa6)

Người bên trái lập phòng xem phim và người bên phải nhập mã phòng và mật khẩu.

![440828424_1493768367907482_1648031953785594319_n](https://github.com/hungq1205/watch-party/assets/84914589/3c5fab71-6328-4172-a814-a7c00894d00d)

Sau đó các người dùng có thể xem phim cùng nhau

Mặc dù cùng phòng nhưng người chủ phòng sẽ có quyền cao hơn như điều chỉnh phim, trong khi người tham gia chỉ có thể request chủ phòng dừng phim

![440813450_1872127009921111_6990629772177419468_n](https://github.com/hungq1205/watch-party/assets/84914589/f7718e27-5a2d-49e7-bc8c-5f52b2ad089f)

Và các người dùng cũng có thể trò chuyện với nhau

![438328747_1426633691557308_4135359435857859331_n](https://github.com/hungq1205/watch-party/assets/84914589/82cf1915-0c98-417f-a2ba-ccf144e6745a)
