$(document).ready(function() {
    $('#register-form').submit(function(event) {
      event.preventDefault(); // Ngăn chặn form tự submit

      // Lấy dữ liệu từ form
      const formData = {
        username: $('input[name="username"]').val(),
        email: $('input[name="email"]').val(),
        password: $('input[name="password"]').val(),
      };

      // Gửi dữ liệu qua AJAX
      $.ajax({
        url: '/auth/register', // Endpoint đăng ký
        type: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(formData),
        success: function(response) {
            showSuccessToast('Register successfully!');
            // Đặt độ trễ 2 giây trước khi chuyển hướng
            setTimeout(function() {
                window.location.href = '/auth/login'; // Chuyển hướng tới trang đăng nhập
            }, 2000); // 2000 milliseconds = 2 giây
        },
        error: function(xhr, status, error) {
            const errorMessage = xhr.responseJSON.message || 'Register failed!';
            showErrorToast(errorMessage);
        }
      });
    });
  });


  $(document).ready(function() {
    // Hàm để gửi yêu cầu đến `/admin/dashboard` với token
    function loadAdminDashboard() {
        const token = localStorage.getItem('token');
        console.log("Token from localStorage:", token);
  
        // Kiểm tra xem token có tồn tại không trước khi gửi yêu cầu
        if (!token) {
            console.error("No token found in localStorage");
            showErrorToast("Authorization token missing!");
            return;
        }
  
        $.ajax({
            url: '/admin/dashboard',
            type: 'GET',
            headers: {
                'Authorization': 'Bearer ' + token
            },
            success: function(response) {
              window.location.href = '/admin/dashboard';
            },
            error: function(xhr, status, error) {
                showErrorToast('Failed to load admin dashboard');
            }
        });
    }
  
    // Xử lý sự kiện khi form đăng nhập được submit
    $('#login-form').submit(function(event) {
        event.preventDefault(); // Ngăn chặn form tự submit
  
        // Lấy dữ liệu từ form
        const formData = {
            email: $('input[name="email"]').val(),
            password: $('input[name="password"]').val(),
        };
  
        // Gửi dữ liệu qua AJAX để đăng nhập
        $.ajax({
            url: '/auth/login', // Endpoint đăng nhập
            type: 'POST',
            contentType: 'application/json',
            data: JSON.stringify(formData),
            success: function(response) {
                console.log("Token received:", response.token);
                console.log("User id:", response.user_id);
                // Lưu token vào localStorage
                localStorage.setItem('token', response.token);
  
                showSuccessToast('Login successfully!');
                // Kiểm tra vai trò người dùng
                if (response.role === 'admin') {
                    // Đặt độ trễ 2 giây rồi gọi loadAdminDashboard để kiểm tra quyền truy cập
                    setTimeout(function() {
                        loadAdminDashboard(); // Gọi hàm để chuyển đến trang admin
                    }, 2000); // 2000 milliseconds = 2 giây
                } else {
                    // Chuyển hướng đến trang /home nếu không phải admin
                    setTimeout(function() {
                        // Truyền user_id qua query string
                        window.location.href = `/home?user_id=${response.user_id}`;
                    }, 10000);
                }
            },
            error: function(xhr, status, error) {
                const errorMessage = xhr.responseJSON.error || 'Login failed!';
                showErrorToast(errorMessage);
            }            
        });
    });
  });
  

// // Log toàn bộ nội dung localStorage
// console.log("Current localStorage data:", localStorage);


window.onload = function() {
  const params = new URLSearchParams(window.location.search);
  const message = params.get("message");
  if (message) {
      // Hiển thị Toast
      showErrorToast(message);
  }
};
