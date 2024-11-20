// <!-- updateField -->
function makeEditableTitle(element, field, id) {
	const currentValue = element.innerText;
	const input = document.createElement('input');
	input.type = 'text';
	input.value = currentValue;
	input.classList.add('form-control', 'form-control-sm');

	// Thêm sự kiện khi người dùng nhấn Enter để lưu dữ liệu
	input.addEventListener('keyup', function(event) {
		if (event.key === 'Enter' || event.keyCode === 13) {
			if (validateTitle(input.value)) { // Validate trước khi lưu
				saveData(input, field, id); // Gọi hàm lưu dữ liệu
			}
		}
	});

	// Khi người dùng nhấp ra ngoài, cũng tự động lưu dữ liệu
	input.addEventListener('blur', function() {
		if (validateTitle(input.value)) { // Validate trước khi lưu
			saveData(input, field, id); // Gọi hàm lưu dữ liệu
		}
	});

	// Thay thế phần text bằng input
	element.innerHTML = '';
	element.appendChild(input);
	input.focus(); // Đưa con trỏ chuột vào input để tiện chỉnh sửa
}

// Hàm validate cho Title
function validateTitle(value) {
	if (!value || value.trim().length < 3 || value.trim().length > 100) {
		// Hiển thị thông báo lỗi qua Toastify
		Toastify({
			text: "Title must be between 3 and 100 characters!",
			duration: 3000, // Thời gian hiển thị 3 giây
			close: true, // Cho phép đóng thông báo
			gravity: "top", // Vị trí hiển thị ở trên
			position: "right", // Vị trí hiển thị bên phải
			backgroundColor: "linear-gradient(to right, #ea0606, #ff667c)", // Màu nền thông báo lỗi
			className: "error-toast" // Lớp CSS tùy chỉnh cho thông báo lỗi
		}).showToast();
		return false;
	}
	return true;
}

function makeEditableDescription(element, field, id) {
	const currentValue = element.innerText;
	const input = document.createElement('textarea');
	input.value = currentValue;
	input.classList.add('form-control', 'form-control-sm');

	input.addEventListener('keyup', function(event) {
		if (event.key === 'Enter' || event.keyCode === 13) {
			if (validateDescription(input.value)) { // Validate trước khi lưu
				saveData(input, field, id);
			}
		}
	});

	input.addEventListener('blur', function() {
		if (validateDescription(input.value)) { // Validate trước khi lưu
			saveData(input, field, id);
		}
	});

	element.innerHTML = '';
	element.appendChild(input);
	input.focus();
}
// Hàm validate cho Description
function validateDescription(value) {
	if (!value || value.trim().length < 3 || value.trim().length > 250) {
		// Hiển thị thông báo lỗi qua Toastify
		Toastify({
			text: "Title must be between 3 and 250 characters!",
			duration: 3000, // Thời gian hiển thị 3 giây
			close: true, // Cho phép đóng thông báo
			gravity: "top", // Vị trí hiển thị ở trên
			position: "right", // Vị trí hiển thị bên phải
			backgroundColor: "linear-gradient(to right, #ea0606, #ff667c)", // Màu nền thông báo lỗi
			className: "error-toast" // Lớp CSS tùy chỉnh cho thông báo lỗi
		}).showToast();
		return false;
	}
	return true;
}

function makeEditableDuration(element, field, id) {
	const currentValue = element.innerText;
	const input = document.createElement('textarea');
	input.value = currentValue;
	input.classList.add('form-control', 'form-control-sm');

	input.addEventListener('keyup', function(event) {
		if (event.key === 'Enter' || event.keyCode === 13) {
			if (validateDuration(input.value)) { // Validate trước khi lưu
				saveData(input, field, id);
			}
		}
	});

	input.addEventListener('blur', function() {
		if (validateDuration(input.value)) { // Validate trước khi lưu
			saveData(input, field, id);
		}
	});

	element.innerHTML = '';
	element.appendChild(input);
	input.focus();
}

// Hàm validate cho Duration
function validateDuration(value) {
	if (!value || value.trim().length < 3 || value.trim().length > 250) {
		// Hiển thị thông báo lỗi qua Toastify
		Toastify({
			text: "Title must be between 3 and 250 characters!",
			duration: 3000, // Thời gian hiển thị 3 giây
			close: true, // Cho phép đóng thông báo
			gravity: "top", // Vị trí hiển thị ở trên
			position: "right", // Vị trí hiển thị bên phải
			backgroundColor: "linear-gradient(to right, #ea0606, #ff667c)", // Màu nền thông báo lỗi
			className: "error-toast" // Lớp CSS tùy chỉnh cho thông báo lỗi
		}).showToast();
		return false;
	}
	return true;
}

function makeEditableSlug(element, field, id) {
	const currentValue = element.innerText;
	const input = document.createElement('input');
	input.type = 'text';
	input.value = currentValue;
	input.classList.add('form-control', 'form-control-sm');

	input.addEventListener('keyup', function(event) {
		if (event.key === 'Enter' || event.keyCode === 13) {
			if (validateSlug(input.value)) { // Validate trước khi lưu
				saveData(input, field, id);
			}
		}
	});

	input.addEventListener('blur', function() {
		if (validateSlug(input.value)) { // Validate trước khi lưu
			saveData(input, field, id);
		}
	});

	element.innerHTML = '';
	element.appendChild(input);
	input.focus();
}

// Hàm validate cho Slug
function validateSlug(value) {
	if (!value || value.trim().length === 0) {
		// Hiển thị thông báo lỗi qua Toastify
		Toastify({
			text: "Slug is required!",
			duration: 3000, // Thời gian hiển thị 3 giây
			close: true, // Cho phép đóng thông báo
			gravity: "top", // Vị trí hiển thị ở trên
			position: "right", // Vị trí hiển thị bên phải
			backgroundColor: "linear-gradient(to right, #ea0606, #ff667c)", // Màu nền thông báo lỗi
			className: "error-toast" // Lớp CSS tùy chỉnh cho thông báo lỗi
		}).showToast();
		return false;
	}
	return true;
}

function makeEditableStatus(element, field, id, currentStatus) {
	const select = document.createElement('select');
	select.classList.add('form-control', 'form-control-sm');

	const option1 = document.createElement('option');
	option1.value = 1;
	option1.text = 'Presently';
	option1.selected = currentStatus === 1;

	const option2 = document.createElement('option');
	option2.value = 2;
	option2.text = 'Hidden';
	option2.selected = currentStatus === 2;

	select.appendChild(option1);
	select.appendChild(option2);

	// Lưu dữ liệu khi người dùng thay đổi lựa chọn
	select.addEventListener('change', function() {
		// Lưu ý thay đổi field thành 'status' thay vì 'Status'
		saveData(select, field, id);
	});

	// Lưu dữ liệu khi nhấp ra ngoài
	select.addEventListener('blur', function() {
		saveData(select, field, id);
	});

	element.innerHTML = '';
	element.appendChild(select);
	select.focus();
}

function makeEditableMaxQuality(element, field, id, currentQuality) {
	const select = document.createElement('select');
	select.classList.add('form-control', 'form-control-sm');

	const qualities = [{
			value: 1,
			text: 'Cam'
		},
		{
			value: 720,
			text: 'HD'
		},
		{
			value: 1080,
			text: 'Full HD'
		},
		{
			value: 1440,
			text: '2K'
		},
		{
			value: 2160,
			text: '4K'
		}
	];

	qualities.forEach(quality => {
		const option = document.createElement('option');
		option.value = quality.value;
		option.text = quality.text;
		option.selected = currentQuality === quality.value;
		select.appendChild(option);
	});

	select.addEventListener('change', function() {
		saveData(select, field, id);
	});

	select.addEventListener('blur', function() {
		saveData(select, field, id);
	});

	element.innerHTML = '';
	element.appendChild(select);
	select.focus();
}

function makeEditableHotMovie(element, field, id, currentHotMovie) {
	const select = document.createElement('select');
	select.classList.add('form-control', 'form-control-sm');

	const option1 = document.createElement('option');
	option1.value = 1;
	option1.text = 'Hot';
	option1.selected = currentHotMovie === 1;

	const option2 = document.createElement('option');
	option2.value = 2;
	option2.text = 'Không';
	option2.selected = currentHotMovie === 2;

	select.appendChild(option1);
	select.appendChild(option2);

	select.addEventListener('change', function() {
		saveData(select, field, id);
	});

	select.addEventListener('blur', function() {
		saveData(select, field, id);
	});

	element.innerHTML = '';
	element.appendChild(select);
	select.focus();
}

function makeEditableYear(element, field, id, currentYear) {
	const select = document.createElement('select');
	select.classList.add('form-control', 'form-control-sm');

	const years = [];
	for (let i = 2000; i <= new Date().getFullYear(); i++) {
		years.push(i);
	}

	years.forEach(year => {
		const option = document.createElement('option');
		option.value = year;
		option.text = year;
		option.selected = currentYear === year;
		select.appendChild(option);
	});

	select.addEventListener('change', function() {
		saveData(select, field, id);
	});

	select.addEventListener('blur', function() {
		saveData(select, field, id);
	});

	element.innerHTML = '';
	element.appendChild(select);
	select.focus();
}

function makeEditableNumofep(element, field, id, currentNumofep) {
	const select = document.createElement('select');
	select.classList.add('form-control', 'form-control-sm');

	const numofep = [];
	for (let i = 1; i <= 30; i++) {
		numofep.push(i);
	}

	numofep.forEach(numofep => {
		const option = document.createElement('option');
		option.value = numofep;
		option.text = numofep;
		option.selected = currentNumofep === numofep;
		select.appendChild(option);
	});

	select.addEventListener('change', function() {
		saveData(select, field, id);
	});

	select.addEventListener('blur', function() {
		saveData(select, field, id);
	});

	element.innerHTML = '';
	element.appendChild(select);
	select.focus();
}

function makeEditableSeason(element, field, id, currentSeason) {
	const select = document.createElement('select');
	select.classList.add('form-control', 'form-control-sm');

	const season = [];
	for (let i = 1; i <= 30; i++) {
		season.push(i);
	}

	season.forEach(season => {
		const option = document.createElement('option');
		option.value = season;
		option.text = season;
		option.selected = currentSeason === season;
		select.appendChild(option);
	});

	select.addEventListener('change', function() {
		saveData(select, field, id);
	});

	select.addEventListener('blur', function() {
		saveData(select, field, id);
	});

	element.innerHTML = '';
	element.appendChild(select);
	select.focus();
}

function makeEditableSub(element) {
	const id = element.getAttribute('data-id');
	let selectedValues = element.getAttribute('data-sub');

	// Parse the selectedValues từ chuỗi JSON
	try {
		selectedValues = JSON.parse(selectedValues);
	} catch (e) {
		console.error('Error parsing selected values:', e);
		selectedValues = [];
	}

	const languages = ['English', 'Vietnamese', 'Chinese', 'Japanese', 'Korean'];

	// Tạo danh sách checkbox và xử lý checked
	let checkboxes = languages.map(lang => {
		const isChecked = selectedValues.includes(lang) ? 'checked' : '';
		return `<label><input type="checkbox" value="${lang}" ${isChecked}> ${lang}</label><br>`;
	}).join('');

	// Thay thế toàn bộ nội dung của `td` bằng checkbox
	const tdElement = element.closest('td');
	tdElement.innerHTML = checkboxes;

	// Xử lý khi nhấn ra ngoài để lưu lại
	document.addEventListener('click', function onClickOutside(event) {
		if (!tdElement.contains(event.target)) {
			submitLanguages(tdElement, 'sub', id, true); // Truyền `true` để xác định là socket
			document.removeEventListener('click', onClickOutside);
		}
	});

	// Xử lý khi nhấn Enter để lưu lại
	tdElement.addEventListener('keydown', function onKeydown(event) {
		if (event.key === 'Enter') {
			submitLanguages(tdElement, 'sub', id, true); // Truyền `true` để xác định là socket
			tdElement.removeEventListener('keydown', onKeydown);
		}
	});
}

// Hàm thu thập các giá trị checkbox đã chọn và gửi dữ liệu
function submitLanguages(element, fieldOrId, idOrUndefined, isSocket = false) {
	// Lấy các checkbox được chọn
	const checkboxes = element.querySelectorAll('input[type="checkbox"]:checked');
	const selectedLanguages = Array.from(checkboxes).map(checkbox => checkbox.value);

	if (isSocket) {
		// Trong trường hợp WebSocket (makeEditableSubSocket), chỉ truyền ID và mảng đã chọn
		saveData(selectedLanguages, fieldOrId, idOrUndefined);
	} else {
		// Trường hợp DOM element (makeEditableSub)
		saveData(selectedLanguages, fieldOrId, idOrUndefined);
	}
}

// Hàm để lưu dữ liệu sau khi chỉnh sửa
function saveData(elementOrArray, fieldOrId, idOrUndefined) {
	let newValue;

	// Kiểm tra nếu `elementOrArray` là một mảng (dành cho trường hợp ngôn ngữ)
	if (Array.isArray(elementOrArray)) {
		newValue = elementOrArray; // Đây là mảng giá trị ngôn ngữ đã chọn
	} else {
		// Xử lý nếu `elementOrArray` là phần tử DOM (giá trị đơn)
		newValue = elementOrArray.value;

		// Nếu cần cập nhật các trường có kiểu số
		if (fieldOrId === 'status' || fieldOrId === 'maxquality' || fieldOrId === 'year' || fieldOrId === 'hotmovie' || fieldOrId === 'numofep' || fieldOrId === 'season') {
			newValue = parseInt(newValue, 10); // Chuyển sang số nguyên
		}
	}

	// Dữ liệu để gửi tới server
	const movieData = {
		field: fieldOrId,
		value: newValue
	};

	// Kiểm tra cú pháp ObjectID
	if (idOrUndefined && idOrUndefined.startsWith("ObjectID(")) {
		idOrUndefined = idOrUndefined.replace(/ObjectID\("(.*)"\)/, "$1"); // Chỉ lấy giá trị bên trong ObjectID
	}

	// Thực hiện cập nhật dữ liệu qua AJAX
	$.ajax({
		url: '/admin/update-movie-field/' + idOrUndefined, // URL cho việc cập nhật
		type: 'POST', // Phương thức POST
		contentType: 'application/json', // Dữ liệu gửi là JSON
		data: JSON.stringify(movieData), // Chuyển đổi dữ liệu thành JSON
		success: function(response) {
			console.log("Success response:", response);

			// Nếu là mảng (ngôn ngữ), cập nhật lại giao diện với các giá trị đã chọn
			if (Array.isArray(newValue)) {
				elementOrArray.innerHTML = newValue.join(', ');
			} else {
				// Nếu là giá trị đơn, cập nhật giao diện với giá trị mới
				elementOrArray.parentElement.innerHTML = newValue;
			}

			showSuccessToast("Movie updated successfully!");
		},
		error: function(xhr, status, error) {
			showErrorToast(xhr.responseJSON.message || "Something went wrong!");
		}
	});
}


// <!-- delete -->
function deleteMovie(id) {
	showOkCancelToast('Are you sure you want to delete this movie?', function() {
		// Kiểm tra nếu ID có cú pháp ObjectID(...) thì loại bỏ phần dư thừa
		if (id.startsWith("ObjectID(")) {
			id = id.replace(/ObjectID\("(.*)"\)/, "$1"); // Chỉ lấy giá trị bên trong dấu ngoặc
		}
		$.ajax({
			url: '/admin/delete-movie/' + id,
			type: 'DELETE',
			success: function(response) {
				showSuccessToast("Movie delete successfully!");
			},
			error: function(xhr, status, error) {
				showErrorToast(xhr.responseJSON.message);
			}
		});
	})
}

// <!-- hien thi update -->
function openUpdatePopup(movie) {
	// Lấy giá trị từ movie object và đổ vào form
	document.getElementById('movieId').value = movie.ID; // ID
	document.getElementById('title').value = movie.Title; // Title
	document.getElementById('name_eng').value = movie.NameEng; // NameEng
	document.getElementById('tags').value = movie.Tags;
	document.getElementById('slug').value = movie.Slug; // Slug
	document.getElementById('description').value = movie.Description; // Description
	document.getElementById('duration').value = movie.Duration; // Duration
	document.getElementById('trailer').value = movie.Trailer; // Trailer

	// Status (Hiện/Ẩn)
	document.getElementById('status').value = movie.Status;

	// Hotmovie (Hot/Không)
	document.getElementById('hotmovie').value = movie.Hotmovie;

	// Max Quality (Cam, HD, Full HD, 2K, 4K)
	document.getElementById('maxquality').value = movie.MaxQuality;

	// Season
	document.getElementById('season').value = movie.Season;

	// Numofep (Số tập)
	document.getElementById('numofep').value = movie.Numofep;

	// Year
	document.getElementById('year').value = movie.Year;

	// Cập nhật các checkbox cho Category
	let categoryCheckboxes = document.querySelectorAll('#category input[type="checkbox"]');
	categoryCheckboxes.forEach((checkbox) => {
		checkbox.checked = movie.Category.includes(checkbox.value);
	});

	// Cập nhật các checkbox cho Genre
	let genreCheckboxes = document.querySelectorAll('#genre input[type="checkbox"]');
	genreCheckboxes.forEach((checkbox) => {
		checkbox.checked = movie.Genre.includes(checkbox.value);
	});

	// Cập nhật country
	document.getElementById('country').value = movie.Country;

	// Cập nhật các checkbox cho Subtitles
	let subCheckboxes = document.querySelectorAll('#sub input[type="checkbox"]');
	subCheckboxes.forEach((checkbox) => {
		checkbox.checked = movie.Sub.includes(checkbox.value);
	});

	// Hiển thị hình ảnh chính
	const imageDiv = document.querySelector('.image');
	imageDiv.innerHTML = `<img src="/admin/assets/uploads/${movie.Image}" alt="Movie Image" class="avatar avatar-sm me-3">`;

	const moreImageDiv = document.querySelector('.moreimage');
	if (Array.isArray(movie.Moreimage) && movie.Moreimage.length > 0) {
		moreImageDiv.innerHTML = movie.Moreimage.map(img =>
			`<img src="/admin/assets/uploads/${img}" alt="More Movie Image" class="avatar avatar-sm me-3" >`
		).join('');
	} else {
		moreImageDiv.innerHTML = "<p>No additional images available.</p>";
	}

	// Cập nhật Image
	document.getElementById('imageaddmovie').value = ""; // Để người dùng có thể upload ảnh mới

	// Cập nhật More Images
	document.getElementById('moreimageaddmovie').value = ""; // Để người dùng có thể upload ảnh mới
	// Hiển thị popup cập nhật
	document.getElementById('updatePopup').style.display = 'flex';
}
// Đóng popup khi nhấn nút "Close"
function openUpdatePopupSocket(button) {
	// Lấy dữ liệu từ các thuộc tính data-* của nút button
	const movie = {
		id: button.getAttribute('data-id'),
		title: button.getAttribute('data-title'),
		name_eng: button.getAttribute('data-name_eng'),
		tags: button.getAttribute('data-tags'),
		slug: button.getAttribute('data-slug'),
		description: button.getAttribute('data-description'),
		duration: button.getAttribute('data-duration'),
		trailer: button.getAttribute('data-trailer'),
		status: button.getAttribute('data-status'),
		hotmovie: button.getAttribute('data-hotmovie'),
		maxquality: button.getAttribute('data-maxquality'),
		season: button.getAttribute('data-season'),
		numofep: button.getAttribute('data-numofep'),
		year: button.getAttribute('data-year'),
		category: button.getAttribute('data-category').split(','), // Chuyển thành mảng
		genre: button.getAttribute('data-genre').split(','), // Chuyển thành mảng
		country: button.getAttribute('data-country'),
		sub: button.getAttribute('data-sub').split(','), // Chuyển thành mảng
		image: button.getAttribute('data-image'),
		moreimage: button.getAttribute('data-moreimage').split(',') // Chuyển thành mảng
	};
	// Đổ dữ liệu vào form
	document.getElementById('movieId').value = movie.id || '';
	document.getElementById('title').value = movie.title || '';
	document.getElementById('name_eng').value = movie.name_eng || '';
	document.getElementById('tags').value = movie.tags || '';
	document.getElementById('slug').value = movie.slug || '';
	document.getElementById('description').value = movie.description || '';
	document.getElementById('duration').value = movie.duration || '';
	document.getElementById('trailer').value = movie.trailer || '';
	document.getElementById('status').value = movie.status || '';
	document.getElementById('hotmovie').value = movie.hotmovie || '';
	document.getElementById('maxquality').value = movie.maxquality || '';
	document.getElementById('season').value = movie.season || '';
	document.getElementById('numofep').value = movie.numofep || '';
	document.getElementById('year').value = movie.year || '';

	// Cập nhật checkbox cho Category
	let categoryCheckboxes = document.querySelectorAll('#category input[type="checkbox"]');
	categoryCheckboxes.forEach((checkbox) => {
		checkbox.checked = movie.category.includes(checkbox.value);
	});

	// Cập nhật checkbox cho Genre
	let genreCheckboxes = document.querySelectorAll('#genre input[type="checkbox"]');
	genreCheckboxes.forEach((checkbox) => {
		checkbox.checked = movie.genre.includes(checkbox.value);
	});

	// Cập nhật country
	document.getElementById('country').value = movie.country || '';

	// Cập nhật checkbox cho Subtitles
	let subCheckboxes = document.querySelectorAll('#sub input[type="checkbox"]');
	subCheckboxes.forEach((checkbox) => {
		checkbox.checked = movie.sub.includes(checkbox.value);
	});

	// Hiển thị hình ảnh chính
	const imageDiv = document.querySelector('.image');
	if (movie.image) {
		imageDiv.innerHTML = `<img src="/uploads/images/${movie.image}" alt="Movie Image" class="avatar me-3">`;
	} else {
		imageDiv.innerHTML = "<p>No main image available.</p>";
	}

	// Hiển thị thêm hình ảnh
	const moreImageDiv = document.querySelector('.moreimage');
	if (movie.moreimage && movie.moreimage.length > 0) {
		moreImageDiv.innerHTML = movie.moreimage.map(img =>
			`<img src="/uploads/images/${img}" alt="More Movie Image" class="avatar me-3" >`
		).join('');
	} else {
		moreImageDiv.innerHTML = "<p>No additional images available.</p>";
	}

	// Reset các trường upload ảnh
	document.getElementById('imageaddmovie').value = ""; // Để người dùng upload ảnh mới
	document.getElementById('moreimageaddmovie').value = ""; // Để người dùng upload thêm ảnh

	// Hiển thị popup
	document.getElementById('updatePopup').style.display = 'flex';
}
// Đóng popup khi nhấn vào overlay bên ngoài
$('#updatePopup .popup__overlay').on('click', function() {
	$('#updatePopup').css('display', 'none');
});

// <!-- add -->
$(document).ready(function() {
	$("#addmovieForm").on("submit", function(e) {
		e.preventDefault(); // Ngăn không cho form gửi trực tiếp

		var formData = new FormData(this); // Tạo đối tượng FormData để gửi dữ liệu, bao gồm cả file
		// console.log(formData.get('name_eng'));
		$.ajax({
			url: "/admin/add-movie", // Route để xử lý
			type: "POST", // Phương thức HTTP
			data: formData, // Dữ liệu form
			processData: false, // Ngăn không cho jQuery xử lý dữ liệu
			contentType: false, // Ngăn jQuery thiết lập kiểu nội dung mặc định
			success: function(response) {
				showSuccessToast("Movie add successfully!");
				document.getElementById("addPopup").style.display = "none";
			},
			error: function(xhr, status, error) {
				showErrorToast(xhr.responseJSON.message);
			}
		});
	});
});

//open popup
const popup = document.getElementById('addPopup');
const openPopupBtn = document.getElementById('openPopupBtn');

// Open popup when "Add Movie" button is clicked
openPopupBtn.addEventListener('click', function() {
	popup.style.display = 'flex'; // Show popup
});

// Đóng popup khi nhấn vào overlay bên ngoài
$('#addPopup .popup__overlay').on('click', function() {
	$('#addPopup').css('display', 'none');
});

// <!-- update -->
$(document).ready(function() {
	$('#updatemovieForm').on('submit', function(e) {
		e.preventDefault();

		var formData = new FormData(this); // Lấy toàn bộ dữ liệu từ form
		var movieId = $('#movieId').val(); // Lấy ID của phim

		$.ajax({
			url: '/admin/update-movie/' + movieId, // Gửi yêu cầu tới endpoint Golang
			type: 'POST',
			data: formData,
			contentType: false,
			processData: false,
			success: function(response) {
				showSuccessToast("Movie update successfully!");
				// Ẩn popup và tải lại trang sau khi cập nhật thành công
				$('#updatePopup').hide();
			},
			error: function(xhr, status, error) {
				showErrorToast(xhr.responseJSON.message);
			}
		});
	});
});

// <!-- Hien thi bang websocket -->
// Kết nối WebSocket
let socket = new WebSocket("ws://localhost:8080/ws");

socket.onmessage = function(event) {
	// Khi nhận được thông báo từ server qua WebSocket
	let message = event.data;
	// console.log("Received message:", message);

	try {
		// Parse message thành JSON object
		let data = JSON.parse(message);

		// Kiểm tra loại thông báo (movie hoặc episode)
		if (data.type === "movie") {
			console.log("Movie update detected");
			updateMoviesSocket(data.movies); // Cập nhật danh sách movies
		} else if (data.type === "episode") {
			console.log("Episode update detected, movie ID:", data.movieID);
			updateEpisodes(data.movieID); // Cập nhật danh sách episodes cho movie tương ứng
		} else if (data.type === "quality") {
            console.log("Quality update detected, movie ID:", data.movieID, "episode ID:", data.episodeID, "server ID:", data.serverID);
            updateQualities(data.movieID, data.episodeID, data.serverID); // Cập nhật danh sách qualities cho episode và server tương ứng
		}
	} catch (error) {
		console.error("Error parsing message:", error);
	}
};
// Hàm updateMoviesSocket chỉ được gọi khi có cập nhật từ WebSocket
function updateMoviesSocket(movies) {
	let moviesList = document.getElementById('movies-list');
	moviesList.innerHTML = ""; // Xóa danh sách cũ

	movies.forEach(movie => {
		addMovieRow(movie); // Hàm tạo và thêm hàng cho từng phim
	});
}
let queryString = window.location.search;
console.log(queryString);
let isMoviesLoaded = false;
function updateMovies() {
	if (isMoviesLoaded) return; // Nếu dữ liệu đã tải, dừng thực thi
	isMoviesLoaded = true;
	fetch('/admin/movies' + queryString)   // Giả sử đây là API trả về danh sách movie
		.then(response => response.json())
		.then(data => {
			let moviesList = document.getElementById('movies-list');
			moviesList.innerHTML = ""; // Xóa danh sách cũ

			data.movies.forEach(movie => {
				addMovieRow(movie); // Hàm tạo và thêm hàng cho từng phim
			});
		})
		.catch(err => {
			console.error("Failed to fetch movies:", err);
		});
}

// Hàm chung để thêm hàng phim
function addMovieRow(movie) {
	let moviesList = document.getElementById('movies-list');
	let qualityText = '';
	switch (movie.MaxQuality) {
		case 1:
			qualityText = 'Cam';
			break;
		case 720:
			qualityText = 'HD';
			break;
		case 1080:
			qualityText = 'Full HD';
			break;
		case 1440:
			qualityText = '2K';
			break;
		case 2160:
			qualityText = '4K';
			break;
	}

	let row = `
                        <tr class="text_span_table1" data-id="${movie.ID}">
                            <td>
                                <div class="d-flex px-2 py-1">
                                    <div>
                                        <img src="/uploads/images/${movie.Image}" class="text_span_table avatar avatar-sm me-3" alt="${movie.Title}">
                                    </div>
                                    <div class="d-flex flex-column justify-content-center">
                                        <h6 class="text_span_table mb-0 text-sm" ondblclick="makeEditableTitle(this, 'title', '${movie.ID}')">${movie.Title}</h6>
                                        <p class="text_span_table text-xs text-secondary mb-0" ondblclick="makeEditableDescription(this, 'description', '${movie.ID}')">${movie.NameEng}</p>
                                    </div>
                                </div>
                            </td>
                            <td class="align-middle text-center">
                                ${movie.Moreimage.map((image) => `
                                    <div class="text_span_table more-image-container" style="position: relative; display: inline-block;">
                                        <img src="/uploads/images/${image}" class="avatar me-3" alt="user">
                                        <span class="close-icon" data-id="${movie.ID}" data-filename="${image}">&times;</span>
                                    </div>
                                `).join('')}
                            </td>
                            <td class="align-middle text-center">
                                <iframe style="border-radius: 10px;"  width="300" height="150" src="${movie.Trailer}" title="YouTube video player" frameborder="0" allowfullscreen></iframe>
                            </td>
                            <td class="align-middle text-center">
                                <span class="text-secondary text-xs font-weight-bold" ondblclick="makeEditableSlug(this, 'slug', '${movie.ID}')">${movie.Slug}</span>
                            </td>
                            <td class="align-middle text-center">
                                <span class="text_span_table badge badge-sm bg-gradient-primary" ondblclick="makeEditableStatus(this, 'status', '${movie.ID}', ${movie.Status})">${movie.Status === 1 ? 'Presently' : 'Hidden'}</span>
                            </td>
                            <td class="align-middle text-center">
                                <span class="text_span_table badge badge-sm bg-gradient-primary" ondblclick="makeEditableMaxQuality(this, 'maxquality', '${movie.ID}', ${movie.MaxQuality})">${qualityText}</span>
                            </td>
                            <td class="align-middle text-center">
                                ${movie.Sub.map(sub => ` <span style=" margin-bottom: 5px;" class="text_span_table badge badge-sm bg-gradient-secondary" 
                                      data-id="${movie.ID}" 
                                      data-sub='${JSON.stringify(movie.Sub)}'
                                      ondblclick="makeEditableSub(this)">
                                    ${sub}
                                </span><br>`).join('')}
                            </td>
                            <td class="align-middle text-center">
                              ${movie.CategoryDetails.length > 0 
                                  ? movie.CategoryDetails.map(c => `<span class="text_span_table badge badge-sm bg-gradient-secondary" style="margin-bottom: 5px;" ondblclick="makeEditableCategory(this, 'category', '${movie.ID}')">${c.Title}</span><br>`).join('') 
                                  : 'Category null'}
                            </td>
                            <td class="align-middle text-center">
                                ${movie.GenreDetails.length > 0 
                                    ? movie.GenreDetails.map(g => `<span class="text_span_table badge badge-sm bg-gradient-secondary" style="margin-bottom: 5px;" ondblclick="makeEditableGenre(this, 'genre', '${movie.ID}')">${g.Title}</span><br>`).join('') 
                                    : 'Genre null'}
                            </td>
                            <td class="align-middle text-center">
                                <span class="text-secondary text-xs font-weight-bold">${movie.CountryDetails ? movie.CountryDetails[0].Title : 'Quốc gia trống'}</span>
                            </td>
                            <td class="align-middle text-center">
                                <span class="text-secondary text-xs font-weight-bold" ondblclick="makeEditableYear(this, 'year', '${movie.ID}', ${movie.Year})">${movie.Year}</span>
                            </td>
                            <td class="align-middle text-center">
                                <span class="text-secondary text-xs font-weight-bold" ondblclick="makeEditableNumofep(this, 'numofep', '${movie.ID}', ${movie.Numofep})">${movie.Numofep}</span>
                            </td>
                            <td class="align-middle text-center">
                                <span class="text-secondary text-xs font-weight-bold" ondblclick="makeEditableSeason(this, 'season', '${movie.ID}', ${movie.Season})">${movie.Season}</span>
                            </td>
                            <td class="align-middle text-center">
                                <span class="text-secondary text-xs font-weight-bold" ondblclick="makeEditableHotMovie(this, 'hotmovie', '${movie.ID}', ${movie.Hotmovie})">${movie.Hotmovie === 1 ? 'Hot' : 'Không'}</span>
                            </td>
                            <td class="align-middle text-center">
                                <span class="text-secondary text-xs font-weight-bold" ondblclick="makeEditableDuration(this, 'duration', '${movie.ID}')">${movie.Duration}</span>
                            </td>
                            <td class="align-middle text-center">
                                <span class="text-secondary text-xs font-weight-bold">${movie.Views}</span>
                            </td>
                            <td class="sticky-col-right align-middle text-center" style="background: #000000;" >
                                <button type="button" class="btn btn-secondary" 
                                    data-id="${movie.ID}"
                                    data-title="${movie.Title}"
                                    data-name_eng="${movie.NameEng}"
                                    data-tags="${movie.Tags}"
                                    data-slug="${movie.Slug}"
                                    data-description="${movie.Description}"
                                    data-duration="${movie.Duration}"
                                    data-trailer="${movie.Trailer}"
                                    data-status="${movie.Status}"
                                    data-hotmovie="${movie.Hotmovie}"
                                    data-maxquality="${movie.MaxQuality}"
                                    data-season="${movie.Season}"
                                    data-numofep="${movie.Numofep}"
                                    data-year="${movie.Year}"
                                    data-category="${movie.Category.join(',')}"
                                    data-genre="${movie.Genre.join(',')}"
                                    data-country="${movie.Country}"
                                    data-sub="${movie.Sub.join(',')}"
                                    data-image="${movie.Image}"
                                    data-moreimage="${movie.Moreimage.join(',')}"
                                    onclick="openUpdatePopupSocket(this)">
                                    <i class="fa fa-edit"></i>
                                </button>
                                <button type="button" class="btn btn-secondary" onclick="deleteMovie('${movie.ID}')"><i class="fa fa-trash"></i></button>
                            </td>
                        </tr>
                    `;
	// Thêm hàng vào bảng
	moviesList.insertAdjacentHTML('beforeend', row);
}

// Lần đầu tải trang, cập nhật danh sách movies
document.addEventListener("DOMContentLoaded", updateMovies);

// <!-- Xóa từng ảnh  -->
$(document).ready(function() {
	// Lắng nghe sự kiện khi nhấn vào biểu tượng "x" để xóa hình ảnh (sử dụng event delegation)
	$(document).on('click', '.close-icon', function() {
		var movieID = $(this).data('id'); // Lấy movie ID từ thuộc tính data-id

		// Kiểm tra nếu ID có cú pháp ObjectID(...) thì loại bỏ phần dư thừa
		if (movieID.startsWith("ObjectID(")) {
			movieID = movieID.replace(/ObjectID\("(.*)"\)/, "$1"); // Chỉ lấy giá trị bên trong dấu ngoặc
		}
		console.log(movieID);
		var filename = $(this).data('filename'); // Lấy tên file từ thuộc tính data-filename
		console.log(filename);

		// Xác nhận trước khi xóa
		showOkCancelToast('Are you sure you want to delete this image?', function() {
			// Gửi yêu cầu AJAX để xóa ảnh
			$.ajax({
				url: '/admin/delete-movie-image', // Đường dẫn route để xử lý xóa ảnh
				type: 'POST', // Phương thức POST
				data: {
					id: movieID, // Gửi movie ID
					filename: filename // Gửi tên ảnh cần xóa
				},
				success: function(response) {
					if (response.success) {
						// Xóa ảnh khỏi giao diện
						showSuccessToast("Image update successfully!");
					}
				},
				error: function(xhr, status, error) {
					// Xử lý lỗi
					showErrorToast(xhr, status, error);
				}
			});
		})
	});
});

// <!-- Xắp xếp vị trí -->
document.addEventListener('DOMContentLoaded', function() {
	var el = document.getElementById('movies-list');
	var sortable = Sortable.create(el, {
		onEnd: function(evt) {
			var movies = [];

			// Lấy danh sách các hàng <tr> trong thứ tự hiện tại
			document.querySelectorAll('#movies-list tr').forEach(function(row, index) {
				// Lấy ID thực từ thuộc tính data-id
				var id = row.getAttribute('data-id');

				// Kiểm tra nếu ID có định dạng "ObjectID(...)", thì chỉ lấy phần giá trị bên trong dấu ngoặc
				if (id.startsWith("ObjectID(")) {
					id = id.replace(/ObjectID\("(.*)"\)/, "$1"); // Chỉ lấy giá trị bên trong
				}

				// Thêm vào mảng movies với vị trí đã cập nhật
				movies.push({
					ID: id,
					Position: index + 1 // Cập nhật vị trí mới dựa trên thứ tự trong danh sách
				});
			});

			// Kiểm tra dữ liệu movies trước khi gửi
			console.log('Dữ liệu movies:', movies);

			// Sử dụng jQuery Ajax để gửi dữ liệu lên server
			$.ajax({
				url: '/admin/movie-update-position', // URL của API cần gửi
				type: 'POST', // Phương thức POST để cập nhật dữ liệu
				contentType: 'application/json', // Loại dữ liệu là JSON
				data: JSON.stringify(movies), // Chuyển đổi mảng movies thành chuỗi JSON
				success: function(response) {
					if (response.error) {
						showErrorToast('Có lỗi xảy ra: ' + response.error);
					} else {
						showSuccessToast("Upadate position successfully!");
					}
				},
				error: function(xhr, status, error) {
					showErrorToast('Có lỗi xảy ra trong quá trình gửi dữ liệu: ' + error);
				}
			});
		}
	});
});

// <!-- socket hien thi episode -->
// Hàm để cập nhật danh sách tập phim cho một movie cụ thể
function updateEpisodes(movieID) {
	if (movieID.startsWith("ObjectID(")) {
		movieID = movieID.replace(/ObjectID\("(.*)"\)/, "$1"); // Chỉ lấy giá trị bên trong
	}
	console.log(movieID)
	// Gửi yêu cầu AJAX để lấy các episode của movie
	// Truyền ID của movie vào thuộc tính data-id của nút #openPopupAddEpisodeBtn
	$('#openPopupAddEpisodeBtn').attr('data-id', movieID);
	// Gửi yêu cầu AJAX để lấy các episode của movie dựa trên movieID
	$.ajax({
		url: `/admin/movies/${movieID}/episodes`, // Thay đổi URL theo route của bạn
		type: 'GET',
		dataType: 'json',
		success: function(response) {
			renderEpisodeTable(response.episodes);
		},
		error: function(xhr, status, error) {
			console.error('Có lỗi xảy ra:', error);
		}
	});
}

// <!-- bam vao de hien episode -->
$(document).ready(function() {
	// Lắng nghe sự kiện nhấp vào mỗi hàng trong bảng movies-list
	$('#movies-list').on('click', 'tr', function() {
		// Lấy ID của movie từ thuộc tính data-id
		let movieID = $(this).data('id');

		if (movieID.startsWith("ObjectID(")) {
			movieID = movieID.replace(/ObjectID\("(.*)"\)/, "$1"); // Chỉ lấy giá trị bên trong
		}

		// console.log(movieID) 
		// Gửi yêu cầu AJAX để lấy các episode của movie
		// Truyền ID của movie vào thuộc tính data-id của nút #openPopupAddEpisodeBtn
		$('#openPopupAddEpisodeBtn').attr('data-id', movieID);

		$.ajax({
			url: `/admin/movies/${movieID}/episodes`, // Thay đổi URL theo route của bạn
			type: 'GET',
			dataType: 'json',
			success: function(response) {
				// Hiển thị danh sách episode trong bảng khác
				renderEpisodeTable(response.episodes);
			},
			error: function(xhr, status, error) {
				console.error('Có lỗi xảy ra:', error);
			}
		});
	});
});

// Hàm hiển thị danh sách episode trong bảng khác
function renderEpisodeTable(episodes) {
	const episodeTable = $('#episode-table-body');
	episodeTable.empty(); // Xóa các hàng cũ
	// console.log(episodes)
	if (episodes === null) {
		// Nếu không có episode nào, hiển thị thông báo "No episode yet"
		episodeTable.append(`
              <tr>
                  <td colspan="4" class="text-center">No episode yet</td>
              </tr>
          `);
		return;
	}
	episodes.forEach(episode => {
		let movieID = episode.movieid
		// console.log(movieID)
		let episodeID = episode._id
		// console.log(episodeID)
		// Tạo các span riêng biệt cho mỗi server title
		let serverText = episode.server_details && episode.server_details.length > 0 ?
			episode.server_details.map(server => `<span id="episode-server-quality" data-serverid=${server._id} data-episodeid=${episodeID} data-movieid=${movieID} class="badge badge-sm bg-gradient-secondary" style="margin-bottom: 5px; cursor: pointer;" >${server.title}</span> <br>`).join(' ') :
			'Danh mục trống';

		const row = `
              <tr>
                <td>
                  <div class="d-flex px-2 py-1">
                    <div>
                      <img src="/uploads/images/${episode.image}" class="avatar avatar-sm me-3" alt="xd">
                    </div>
                    <div class="d-flex flex-column justify-content-center">
                      <h6 class="mb-0 text-sm">Tập: ${episode.number}</h6>
                    </div>
                  </div>
                </td>
                <td class="align-middle text-center text-sm">
                  ${serverText}
                </td>
                <td class="align-middle text-center text-sm">
                  <span class="text-xs font-weight-bold"> ${episode.status === 1 ? 'Presently' : 'Hidden'} </span>
                </td>
                <td class="align-middle text-center">
                  <button type="button" class="btn btn-secondary"
                  data-id="${episode._id}"
                  data-movieid="${episode.movieid}"
                  data-number="${episode.number}"
                  data-image="${episode.image}"
                  data-status="${episode.status}"
                  data-server="${episode.server.join(',')}"
                  onclick="openUpdatePopupEpisode(this)"><i class="fa fa-edit"></i></button>
                  <button type="button" class="btn btn-secondary" onclick="deleteEpisode('${episode._id}')"><i class="fa fa-trash"></i></button>
                </td> 
              </tr>
          `;
		episodeTable.append(row);
	});
}

// <!-- Hien thi popupaddepisode -->
$(document).ready(function() {
	// Lắng nghe sự kiện nhấp vào nút #openPopupAddEpisodeBtn
	$('#openPopupAddEpisodeBtn').on('click', function() {
		// Lấy data-id từ thuộc tính của nút và truyền vào input ẩn trong form
		let movieID = $(this).data('id');
		console.log(movieID)

		// Kiểm tra xem data-id có tồn tại không
		if (!movieID) {
			showErrorToast("Movie ID not selected. Please select a movie before adding episode!");
			return; // Ngừng thực hiện nếu không có data-id
		}

		$('#movieIdAddEpisode').val(movieID);
		// Hiển thị popup
		$('#addEpisodePopup').css('display', 'flex');
	});

	// Đóng popup khi nhấn nút đóng
	$('#closeAddEpisodePopupBtn').on('click', function() {
		$('#addEpisodePopup').css('display', 'none');
	});

	// Đóng popup khi nhấn vào overlay bên ngoài
	$('#addEpisodePopup .popup__overlay').on('click', function() {
		$('#addEpisodePopup').css('display', 'none');
	});
});

//  <!-- add episode -->
$(document).ready(function() {
	$("#addepisodeForm").on("submit", function(e) {
		e.preventDefault(); // Ngăn không cho form gửi trực tiếp

		var formData = new FormData(this); // Tạo đối tượng FormData để gửi dữ liệu, bao gồm cả file
		// console.log(formData.get('name_eng'));
		$.ajax({
			url: "/admin/add-episode", // Route để xử lý
			type: "POST", // Phương thức HTTP
			data: formData, // Dữ liệu form
			processData: false, // Ngăn không cho jQuery xử lý dữ liệu
			contentType: false, // Ngăn jQuery thiết lập kiểu nội dung mặc định
			success: function(response) {
				showSuccessToast("Episode add successfully!");
				document.getElementById("addEpisodePopup").style.display = "none";
			},
			error: function(xhr, status, error) {
				showErrorToast(xhr.responseJSON.message);
			}
		});
	});
});

//  <!-- Hien thi popupupdateepisode -->
function openUpdatePopupEpisode(button) {
	// Lấy dữ liệu từ các thuộc tính data-* của button
	const episodeId = button.getAttribute("data-id");
	const movieId = button.getAttribute("data-movieid");
	const episodeNumber = button.getAttribute("data-number");
	const episodeImage = button.getAttribute("data-image");
	const episodeStatus = button.getAttribute("data-status");
	const episodeServers = button.getAttribute("data-server").split(',');

	// Gán dữ liệu vào các trường của form trong popup
	document.getElementById("IdEpisode").value = episodeId;
	document.getElementById("movieIdUpdateEpisode").value = movieId;
	document.getElementById("number").value = episodeNumber;
	document.getElementById("episodestatus").value = episodeStatus;

	// Đánh dấu các checkbox server
	const serverCheckboxes = document.querySelectorAll("#server input[type='checkbox']");
	serverCheckboxes.forEach(checkbox => {
		checkbox.checked = episodeServers.includes(checkbox.value);
	});

	// Hiển thị popup
	document.getElementById("updateEpisodePopup").style.display = "flex";

	// Đóng popup khi nhấn nút đóng
	$('#closeUpdateEpisodePopupBtn').on('click', function() {
		$('#updateEpisodePopup').css('display', 'none');
	});

	// Đóng popup khi nhấn vào overlay bên ngoài
	$('#updateEpisodePopup .popup__overlay').on('click', function() {
		$('#updateEpisodePopup').css('display', 'none');
	});
}

// <!-- update episode -->
document.getElementById("updateepisodeForm").addEventListener("submit", function(event) {
	event.preventDefault(); // Ngăn chặn form submit truyền thống

	// Lấy dữ liệu từ form
	const formData = new FormData(this);

	// Gửi yêu cầu AJAX
	fetch("/admin/update-episode", {
			method: "POST",
			body: formData,
		})
		.then(response => response.json())
		.then(data => {
			showSuccessToast(data.message);
			document.getElementById("updateEpisodePopup").style.display = "none";
		})
		.catch(error => {
			console.error("Error:", error);
			showErrorToast("An error occurred while updating the episode.");
		});
});

//delete episode 
function deleteEpisode(id) {
	// console.log(id)
	showOkCancelToast('Are you sure you want to delete this episode?', function() {
		$.ajax({
			url: '/admin/delete-episode/' + id,
			type: 'DELETE',
			success: function(response) {
				showSuccessToast("Episode delete successfully!");
			},
			error: function(xhr, status, error) {
				showErrorToast(xhr.responseJSON.message);
			}
		});
	})
}

// <!-- bam vao de hien quality -->
$(document).ready(function() {
	// Lắng nghe sự kiện nhấp vào thẻ span với id là episode-server-quality
	$(document).on('click', '#episode-server-quality', function() {
		// Lấy episodeID và movieID từ thuộc tính data của span
		let episodeID = $(this).data('episodeid');
		// console.log(episodeID)
		let movieID = $(this).data('movieid');
		// console.log(movieID)
		let serverID = $(this).data('serverid');
		// console.log(serverID)

		// Cập nhật ID của movie vào thuộc tính data-id của nút #openPopupAddEpisodeBtn
		$('#openPopupAddQualityBtn').attr('data-episodeid', episodeID);
		$('#openPopupAddQualityBtn').attr('data-movieid', movieID);
		$('#openPopupAddQualityBtn').attr('data-serverid', serverID);

		// Gửi yêu cầu AJAX để lấy dữ liệu dựa trên episodeID và movieID
		$.ajax({
		    url: `/admin/movies/${movieID}/episodes/${episodeID}/server/${serverID}/qualities`, // Thay đổi URL theo route của bạn
		    type: 'GET',
		    dataType: 'json',
		    success: function (response) {
		        // Xử lý dữ liệu episode nhận được (ví dụ: hiển thị thông tin chi tiết của episode)
		        renderQuality(response.qualities);
		    },
		    error: function (xhr, status, error) {
		        console.error('Có lỗi xảy ra:', error);
		    }
		});
	});
});

// <!-- socket hien thi quality -->
// Hàm để cập nhật danh sách tập phim cho một movie cụ thể
function updateQualities(movieID, episodeID, serverID) {
	console.log(movieID)
	console.log(episodeID)
	console.log(serverID)
	// Gửi yêu cầu AJAX để lấy các quality của movie
	// Truyền ID của movie vào thuộc tính data-id của nút #openPopupAddQualityBtn
	// Cập nhật ID của movie vào thuộc tính data-id của nút #openPopupAddEpisodeBtn
    $('#openPopupAddQualityBtn').attr('data-episodeid', episodeID);
    $('#openPopupAddQualityBtn').attr('data-movieid', movieID);
    $('#openPopupAddQualityBtn').attr('data-serverid', serverID);

	// Gửi yêu cầu AJAX để lấy dữ liệu dựa trên episodeID và movieID
    $.ajax({
        url: `/admin/movies/${movieID}/episodes/${episodeID}/server/${serverID}/qualities`, // Thay đổi URL theo route của bạn
        type: 'GET',
        dataType: 'json',
        success: function (response) {
            // Xử lý dữ liệu episode nhận được (ví dụ: hiển thị thông tin chi tiết của episode)
            renderQuality(response.qualities);
        },
        error: function (xhr, status, error) {
            console.error('Có lỗi xảy ra:', error);
        }
    });
}

// Hàm hiển thị danh sách quality trong bảng khác
function renderQuality(qualities) {
	const qualityDiv = $('#quality-body');
	qualityDiv.empty(); // Xóa các hàng cũ
	// console.log(qualities)
	if (qualities === null) {
		// Nếu không có quality nào, hiển thị thông báo "No quality yet"
		qualityDiv.append(`
                  <div class="timeline-block mb-3" style="text-align: center;">No quality yet</div>
          `);
		return;
	}
	qualities.forEach(quality => {
		const row = `
              <div class="timeline-block mb-3">
              <span class="timeline-step">
                <span ondblclick="makeEditQualityTitle(this, 'title', '${quality._id}', '${quality.title}', '${quality.movie_id}', '${quality.episode_id}', '${quality.server_id}')" class="badge badge-sm 
				${quality.title === 'CAM' ? 'bg-gradient-success' :
				quality.title === 'HD' ? 'bg-gradient-info' :
				quality.title === 'FULL HD' ? 'bg-gradient-warning' :
				quality.title === '2K' ? 'bg-gradient-primary' :
				quality.title === '4K' ? 'bg-gradient-dark' : ''}">
				${quality.title}
				</span>
              </span>
              <div class="timeline-content">
			 	<h6 ondblclick="makeEditQualityDescription(this, 'description', '${quality._id}', '${quality.movie_id}', '${quality.episode_id}', '${quality.server_id}')"  style="text-align: center;" class="text-dark text-sm font-weight-bold mb-4">${quality.description}</h6>
			  	<div style="margin-bottom: 20px;display: flex;justify-content: space-between;align-items: center;" >
					<span  ondblclick="makeEditQualityStatus(this, 'status', '${quality._id}', ${quality.status}, '${quality.movie_id}', '${quality.episode_id}', '${quality.server_id}')"  class="text-xs font-weight-bold"> ${quality.status === 1 ? 'Presently' : 'Hidden'} </span>
					<button type="button" style="margin-bottom: 0;" class="btn btn-dark" onclick="deleteQuality('${quality._id}')"><i class="fa fa-trash"></i></button>
				</div>
                <div style="position: relative; width: 100%; height: 180px;">
					<iframe width="100%" height="180" src="${quality.videourl}" title="YouTube video player" 
						frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; 
						gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" 
						allowfullscreen></iframe>
					
					<!-- Nút chỉnh sửa nằm trên iframe -->
					<button class="btn btn-primary" onclick="makeEditQualityVideoUrl(this.parentNode, 'videourl', '${quality._id}', '${quality.movie_id}', '${quality.episode_id}', '${quality.server_id}')"
						style="position: absolute; top: 5px; right: 5px; z-index: 10;" class="btn btn-sm btn-primary">
						Chỉnh sửa URL
					</button>
				</div>
              </div>
            </div>
          `;
		  qualityDiv.append(row);
	});
}

// <!-- Hien thi popupaddquality -->
$(document).ready(function() {
	// Lắng nghe sự kiện nhấp vào nút #openPopupAddQualityBtn
	$('#openPopupAddQualityBtn').on('click', function() {
		// Lấy data-id từ thuộc tính của nút và truyền vào input ẩn trong form
		let serverID = $(this).attr('data-serverid');
		let episodeID = $(this).attr('data-episodeid');
		let movieID = $(this).attr('data-movieid');
		console.log("moveid", movieID)
		console.log("episodeID", episodeID)
		console.log("serverID", serverID)

		// Kiểm tra xem data-id có tồn tại không	
		if (!serverID) {
			showErrorToast("Server not selected. Please select a server before adding quality!");
			return; // Ngừng thực hiện nếu không có data-id
		}

		$('#movieIdAddQuality').val(movieID);
		$('#episodeIdAddQuality').val(episodeID);
		$('#serverIdAddQuality').val(serverID);
		// Hiển thị popup
		$('#addQualityPopup').css('display', 'flex');
	});

	 // Đóng popup khi nhấn vào overlay bên ngoài
	 $('#addQualityPopup .popup__overlay').on('click', function() {
        // Đóng popup
        $('#addQualityPopup').css('display', 'none');
    });
});

//add quality 
$(document).ready(function() {
	$("#addqualityForm").on("submit", function(e) {
		e.preventDefault(); // Ngăn không cho form gửi trực tiếp

		var formData = new FormData(this); // Tạo đối tượng FormData để gửi dữ liệu, bao gồm cả file
		// console.log(formData.get('name_eng'));
		$.ajax({
			url: "/admin/add-quality", // Route để xử lý
			type: "POST", // Phương thức HTTP
			data: formData, // Dữ liệu form
			processData: false, // Ngăn không cho jQuery xử lý dữ liệu
			contentType: false, // Ngăn jQuery thiết lập kiểu nội dung mặc định
			success: function(response) {
				showSuccessToast("Quality add successfully!");
				document.getElementById("addQualityPopup").style.display = "none";
			},
			error: function(xhr, status, error) {
				showErrorToast(xhr.responseJSON.message);
			}
		});
	});
});


function makeEditQualityTitle(element, field, qualityId, currentQualityTitle, movieId, episodeId, serverId) {
    const select = document.createElement('select');
    select.classList.add('form-control', 'form-control-sm');
    select.style.width = '80px';

    const resolutions = ['CAM', 'HD', 'FULL HD', '2K', '4K'];

    resolutions.forEach(resolution => {
        const option = document.createElement('option');
        option.value = resolution;
        option.text = resolution;
        option.selected = currentQualityTitle === resolution;
        select.appendChild(option);
    });

    select.addEventListener('change', function() {
        saveDataQuality(select, field, qualityId, movieId, episodeId, serverId);
    });

    select.addEventListener('blur', function() {
        saveDataQuality(select, field, qualityId, movieId, episodeId, serverId);
    });

    element.innerHTML = '';
    element.appendChild(select);
    select.focus();
}

function makeEditQualityStatus(element, field, qualityId, currentStatus, movieId, episodeId, serverId) {
    const select = document.createElement('select');
    select.classList.add('form-control', 'form-control-sm');

    const option1 = document.createElement('option');
    option1.value = 1;
    option1.text = 'Presently';
    option1.selected = currentStatus === 1;

    const option2 = document.createElement('option');
    option2.value = 2;
    option2.text = 'Hidden';
    option2.selected = currentStatus === 2;

    select.appendChild(option1);
    select.appendChild(option2);

    select.addEventListener('change', function() {
        saveDataQuality(select, field, qualityId, movieId, episodeId, serverId);
    });

    select.addEventListener('blur', function() {
        saveDataQuality(select, field, qualityId, movieId, episodeId, serverId);
    });

    element.innerHTML = '';
    element.appendChild(select);
    select.focus();
}

function makeEditQualityDescription(element, field, qualityId, movieId, episodeId, serverId) {
    const currentValue = element.innerText;
    const input = document.createElement('textarea');
    input.value = currentValue;
    input.classList.add('form-control', 'form-control-sm');

    input.addEventListener('keyup', function(event) {
        if (event.key === 'Enter' || event.keyCode === 13) {
            if (validateDescription(input.value)) {
                saveDataQuality(input, field, qualityId, movieId, episodeId, serverId);
            }
        }
    });

    input.addEventListener('blur', function() {
        if (validateDescription(input.value)) {
            saveDataQuality(input, field, qualityId, movieId, episodeId, serverId);
        }
    });

    element.innerHTML = '';
    element.appendChild(input);
    input.focus();
}

function makeEditQualityVideoUrl(container, field, qualityId, movieId, episodeId, serverId) {
    const iframe = container.querySelector('iframe');
    const currentValue = iframe.src;

    const input = document.createElement('input');
    input.value = currentValue;
    input.classList.add('form-control', 'form-control-sm');

    input.addEventListener('keyup', function(event) {
        if (event.key === 'Enter' || event.keyCode === 13) {
            if (validateDescription(input.value)) {
                saveDataQuality(input, field, qualityId, movieId, episodeId, serverId);
            }
        }
    });

    input.addEventListener('blur', function() {
        if (validateDescription(input.value)) {
            saveDataQuality(input, field, qualityId, movieId, episodeId, serverId);
        }
    });

    container.innerHTML = ''; // Xóa nội dung cũ
    container.appendChild(input);
    input.focus();
}

function saveDataQuality(elementOrArray, field, qualityId, movieId, episodeId, serverId) {
    let newValue;

    // Xử lý nếu `elementOrArray` là phần tử DOM
    if (!Array.isArray(elementOrArray)) {
        newValue = elementOrArray.value;
    } else {
        newValue = elementOrArray; // Đây là mảng giá trị nếu có (ví dụ: ngôn ngữ)
    }

    // Dữ liệu để gửi tới server
    const qualityData = {
        field: field,
        value: newValue,
        qualityId: qualityId,
        movieId: movieId,
        episodeId: episodeId,
        serverId: serverId
    };

    // Thực hiện cập nhật dữ liệu qua AJAX
    $.ajax({
        url: '/admin/update-qulity-field/' + qualityId, // URL cho việc cập nhật
        type: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(qualityData),
        success: function(response) {
            console.log("Success response:", response);

            // Cập nhật giao diện với giá trị mới
            if (!Array.isArray(newValue)) {
                elementOrArray.parentElement.innerHTML = newValue;
            }

            showSuccessToast("Quality updated successfully!");
        },
        error: function(xhr, status, error) {
            showErrorToast(xhr.responseJSON.message || "Something went wrong!");
        }
    });
}

//delete quality 
function deleteQuality(id) {
	// console.log(id)
	showOkCancelToast('Are you sure you want to delete this quality?', function() {
		$.ajax({
			url: '/admin/delete-quality/' + id,
			type: 'DELETE',
			success: function(response) {
				showSuccessToast("Quality delete successfully!");
			},
			error: function(xhr, status, error) {
				showErrorToast(xhr.responseJSON.message);
			}
		});
	})
}
