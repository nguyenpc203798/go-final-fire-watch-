// Hàm để cập nhật danh sách tập phim cho một movie cụ thể
function getMovies() {
	// Gửi yêu cầu AJAX để lấy dữ liệu 
    $.ajax({
        url: `/movies`, // Thay đổi URL theo route của bạn
        type: 'GET',
        dataType: 'json',
        success: function (response) {
            // Xử lý dữ liệu episode nhận được (ví dụ: hiển thị thông tin chi tiết của episode)
            renderMovie(response.movies);
        },
        error: function (xhr, status, error) {
            console.error('Có lỗi xảy ra:', error);
        }
    });
}

// Hàm hiển thị danh sách movie trong phần tử HTML
function renderMovie(movies) {
   console.log(movies)
   const movieDiv = $('#all-movie-list');
   movieDiv.empty(); // Xóa nội dung cũ

   if (!movies || movies.length === 0) {
       // Nếu không có movie nào, hiển thị thông báo "No movie yet"
       movieDiv.append(`<div style="text-align: center;">No movie yet</div>`);
       return;
   }

   // Render từng movie
   movies.forEach(movie => {
      let qualityText = '';
      switch (movie.MaxQuality) {
         case 1:
            qualityText = 'CAM';
            break;
         case 720:
            qualityText = 'HD';
            break;
         case 1080:
            qualityText = 'FHD';
            break;
         case 1440:
            qualityText = '2K';
            break;
         case 2160:
            qualityText = '4K';
            break;
      }
       const row = `
           <a href="/movies/${movie.ID}" class="movie-item col-3-5 m-5 s-11 to-top show-on-scroll">
            <div>
                <img src="/uploads/images/${movie.Image}" alt="${movie.Title}">
                <div class="movie-item-content">
                    <div class="movie-item-title">
                        ${movie.Title}
                    </div>
                    <div class="movies-infors-card">
                        <div class="movies-infor">
                            <ion-icon name="bookmark-outline"></ion-icon>
                            <span>${movie.Rating || 'N/A'}</span>
                        </div>
                        <div class="movies-infor">
                            <ion-icon name="time-outline"></ion-icon>
                            <span>${movie.Duration || 'N/A'} mins</span>
                        </div>
                        <div class="movies-infor">
                            <ion-icon name="cube-outline"></ion-icon>
                            <span>${qualityText || 'N/A'}</span>
                        </div>
                    </div>
                </div>
            </div>
            <div class="movie-item-overlay"></div>
            <div class="movie-item-act">
                <i class='bx bxs-right-arrow'></i>
                <div>
                    <i class='bx bxs-share-alt'></i>
                    <i class='bx bxs-heart'></i>
                    <i class='bx bx-plus-medical'></i>
                </div>
            </div>
        </a>
       `;
       movieDiv.append(row);
   });

   // Sau khi thêm các thẻ movie-item
   el_to_show = document.querySelectorAll('.show-on-scroll'); // Cập nhật danh sách
   loop(); // Gọi lại loop để kiểm tra các phần tử mới
}







// Hàm để cập nhật danh sách tập phim cho một movie cụ thể
function getcategoriesWithMovies() {
	// Gửi yêu cầu AJAX để lấy dữ liệu 
    $.ajax({
        url: `/categories-movies`, // Thay đổi URL theo route của bạn
        type: 'GET',
        dataType: 'json',
        success: function (response) {
            // Xử lý dữ liệu episode nhận được (ví dụ: hiển thị thông tin chi tiết của episode)
            renderMovie(response.categorieswithmovie);
        },
        error: function (xhr, status, error) {
            console.error('Có lỗi xảy ra:', error);
        }
    });
}

// Hàm hiển thị danh sách movie trong phần tử HTML
function rendercategoriesWithMovies(categorieswithmovie) {
   console.log(categorieswithmovie)
//    const movieDiv = $('#all-movie-list');
//    movieDiv.empty(); // Xóa nội dung cũ

//    if (!categorieswithmovie || categorieswithmovie.length === 0) {
//        // Nếu không có movie nào, hiển thị thông báo "No movie yet"
//        movieDiv.append(`<div style="text-align: center;">No movie yet</div>`);
//        return;
//    }

   // Render từng movie
//    categorieswithmovie.forEach(moviecate => {
//       let qualityText = '';
//       switch (moviecate.MaxQuality) {
//          case 1:
//             qualityText = 'CAM';
//             break;
//          case 720:
//             qualityText = 'HD';
//             break;
//          case 1080:
//             qualityText = 'FHD';
//             break;
//          case 1440:
//             qualityText = '2K';
//             break;
//          case 2160:
//             qualityText = '4K';
//             break;
//       }
//        const row = `
           
//        `;
//        movieDiv.append(row);
//    });

   // Sau khi thêm các thẻ movie-item
//    el_to_show = document.querySelectorAll('.show-on-scroll'); // Cập nhật danh sách
//    loop(); // Gọi lại loop để kiểm tra các phần tử mới
}


$(document).ready(function () {
    getMovies(); // Đổ vào container #movies-slider
    getcategoriesWithMovies(); // Đổ vào container #categories-slider
});