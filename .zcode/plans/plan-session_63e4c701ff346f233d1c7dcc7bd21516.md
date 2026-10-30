Cải thiện toàn bộ giao diện hiện hữu của StreamChat theo hướng thân thiện, rõ ràng và nhất quán, giữ nguyên chức năng, nội dung sản phẩm và nhận diện đang có. Kế hoạch dựa trên code/CSS hiện tại; context loader và bản UI render chưa thể kiểm tra trong plan mode.

1. **Rà soát nền hiện hữu khi bắt đầu thực thi**
   - Chạy Impeccable context đúng target, đọc `reference/operate.md` và `reference/craft-floor.md` trước khi sửa UI.
   - Kiểm tra các màn auth, chat, CSS, trạng thái và chạy app/screenshot hiện có để xác nhận visual baseline; không tạo Product/Design spec mới hay thay visual world.

2. **Làm rõ cấu trúc và thao tác chat**
   - Tổ chức lại sidebar cuộc trò chuyện, tạo chat, header, vùng message và composer theo thứ tự ưu tiên sử dụng; tăng khả năng quét nhanh tên chat, loại chat, nội dung và trạng thái.
   - Làm rõ hierarchy/contrast, khoảng cách, typography, focus/hover/disabled states; giữ palette hiện tại làm nền và chỉ tinh chỉnh khi code/render xác nhận cần thiết.
   - Trình bày message và media-processing states gọn, dễ hiểu; video player có kích thước ổn định, không autoplay.

3. **Hoàn thiện responsive và trạng thái**
   - Trên mobile, thay sidebar bị ẩn bằng cách điều hướng có thể dùng để mở danh sách chat và quay lại conversation; đảm bảo composer, upload, lỗi và nút gửi dùng được trên màn nhỏ/keyboard.
   - Thiết kế trạng thái loading, empty, error, unread, video queued/processing/failed/completed và retry rõ ràng; giữ nguyên hợp đồng API và hành vi gửi tin.
   - Bảo đảm accessibility cơ bản: semantic controls, label, keyboard navigation, visible focus, reduced-motion và contrast phù hợp.

4. **Xác minh**
   - Bổ sung/cập nhật component tests cho luồng responsive/navigation và trạng thái người dùng nếu harness hiện có hỗ trợ.
   - Chạy frontend tests/build, mở app và kiểm tra trực quan ở viewport desktop/mobile; chạy Impeccable detector trên các target đã đổi, xử lý findings liên quan.
   - Chỉ sửa frontend/style/tests cần thiết; không thay đổi backend, API, nội dung sản phẩm hay dependency nếu không có nhu cầu được xác nhận.