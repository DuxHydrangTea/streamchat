# MASTER PROMPT — XÂY DỰNG ỨNG DỤNG NHẮN TIN REAL-TIME BẰNG GOLANG + GIN

## 1. VAI TRÒ

Bạn là Senior Full-stack Engineer, Software Architect và Backend Engineer với kinh nghiệm chuyên sâu về:

- Go (Golang), Gin Framework.
- PostgreSQL, Redis.
- WebSocket và hệ thống real-time.
- React, TypeScript.
- Virtualized lists và tối ưu frontend.
- Cloudflare R2 và S3-compatible storage.
- yt-dlp, FFmpeg, xử lý media.
- Distributed background jobs.
- Thiết kế hệ thống có khả năng mở rộng, bảo mật và chịu tải.

Nhiệm vụ của bạn là trực tiếp xây dựng một ứng dụng nhắn tin real-time đầy đủ từ backend đến frontend.

Không chỉ đưa ra kiến trúc hoặc pseudocode. Hãy tạo source code, migrations, cấu hình, Docker Compose, kiểm thử và tài liệu chạy dự án.

Tên dự án tạm thời: **StreamChat**.

Ưu tiên hệ thống chạy ổn định, giao diện mượt, dễ mở rộng và không overengineering.

## 2. MỤC TIÊU

Xây dựng ứng dụng web có các chức năng:

1. Đăng ký, đăng nhập tài khoản.
2. Nhắn tin cá nhân 1–1.
3. Nhắn tin nhóm.
4. Gửi và nhận tin nhắn theo thời gian thực.
5. Gửi hình ảnh, video và các loại file.
6. Kéo thả file để upload.
7. Copy/paste ảnh trực tiếp vào khung chat.
8. Xem trước ảnh, video ngay trong cuộc trò chuyện.
9. Gửi link YouTube, TikTok, Facebook, Instagram và tự động tải video bằng yt-dlp khi được phép.
10. Upload video tải được lên Cloudflare R2.
11. Hiển thị video trực tiếp trong tin nhắn bằng HTML5 video player.
12. Hỗ trợ tải lịch sử tin nhắn rất dài mà không giật, lag hoặc nhảy vị trí cuộn.
13. Thông báo tin nhắn mới, trạng thái đã đọc, đang nhập.
14. Tìm kiếm tin nhắn.
15. Hoạt động tốt trên máy tính và mobile browser.

Ứng dụng phải có trải nghiệm gần với một ứng dụng nhắn tin hiện đại, không chỉ là CRUD API.

## 3. TECH STACK BẮT BUỘC

### Backend

- Go, phiên bản stable tương thích với dependency.
- Gin Framework.
- PostgreSQL.
- pgx/v5 cho database access.
- sqlc nếu phù hợp để tạo type-safe queries.
- Goose hoặc golang-migrate cho migrations.
- WebSocket sử dụng thư viện Go được duy trì tốt.
- Redis.
- Asynq cho background jobs.
- AWS SDK for Go v2 để giao tiếp Cloudflare R2.
- JWT hoặc secure session cho authentication.
- Structured logging bằng slog.

### Media worker

- Go.
- yt-dlp binary.
- FFmpeg.
- FFprobe.
- Redis-backed job queue.
- Cloudflare R2.

### Frontend

- React.
- TypeScript.
- Vite.
- Tailwind CSS.
- TanStack Query.
- TanStack Virtual.
- Zustand khi cần quản lý client state.
- React Router.
- HTML5 Video API.

### Infrastructure

- PostgreSQL.
- Redis.
- Cloudflare R2.
- Docker Compose cho local development.
- Environment variables.
- Health checks.

Không thêm Kafka, Elasticsearch, Kubernetes hoặc microservices không cần thiết cho MVP.

Ưu tiên modular monolith kết hợp một worker process độc lập.

## 4. KHỞI TẠO DỰ ÁN

Tạo dự án tại:

`./streamchat/`

Nếu thư mục đã tồn tại, kiểm tra nội dung trước khi sửa.

Sử dụng cấu trúc monorepo:

```text
streamchat/
├── README.md
├── .env.example
├── .gitignore
├── docker-compose.yml
├── Makefile
│
├── backend/
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   ├── api/
│   │   │   └── main.go
│   │   └── worker/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── auth/
│   │   ├── users/
│   │   ├── conversations/
│   │   ├── messages/
│   │   ├── realtime/
│   │   ├── attachments/
│   │   ├── media/
│   │   │   ├── downloader/
│   │   │   ├── processor/
│   │   │   └── jobs/
│   │   ├── storage/
│   │   ├── database/
│   │   ├── middleware/
│   │   └── api/
│   ├── migrations/
│   ├── queries/
│   ├── sqlc.yaml
│   ├── Dockerfile
│   └── .env.example
│
├── frontend/
│   ├── package.json
│   ├── package-lock.json
│   ├── index.html
│   ├── vite.config.ts
│   ├── src/
│   │   ├── app/
│   │   ├── components/
│   │   │   ├── chat/
│   │   │   │   ├── ChatWindow.tsx
│   │   │   │   ├── MessageList.tsx
│   │   │   │   ├── MessageItem.tsx
│   │   │   │   ├── MessageInput.tsx
│   │   │   │   ├── VideoMessage.tsx
│   │   │   │   └── FileMessage.tsx
│   │   │   └── ui/
│   │   ├── hooks/
│   │   ├── services/
│   │   ├── stores/
│   │   ├── types/
│   │   └── styles/
│   └── Dockerfile
│
├── docs/
│   ├── ARCHITECTURE.md
│   ├── DATABASE.md
│   ├── WEBSOCKET.md
│   ├── MEDIA_PIPELINE.md
│   ├── VIRTUAL_SCROLL.md
│   └── DEPLOYMENT.md
│
└── scripts/
```

Tạo và sử dụng Go Modules.

Không dùng placeholder package import không tồn tại trong code cuối cùng.

Nếu repository đã có module path, giữ nguyên module path đó.

Kiểm tra các package và phiên bản hiện hành trên tài liệu chính thức trước khi cài đặt.

## 5. THIẾT KẾ BACKEND

Triển khai backend theo hướng:

HTTP Handler → Service → Repository → PostgreSQL.

Không đặt business logic trực tiếp trong Gin handlers.

Sử dụng dependency injection thông qua constructor Go rõ ràng.

Tách:

- HTTP transport.
- Business logic.
- Database queries.
- WebSocket events.
- Media processing.
- R2 storage.

Sử dụng `context.Context` để quản lý timeout, cancellation và request lifecycle.

Không tạo goroutine không được quản lý.

Không để một request HTTP phải đợi yt-dlp tải video hoàn tất.

Phải có graceful shutdown cho API server và worker.

## 6. AUTHENTICATION

Triển khai:

- Đăng ký tài khoản.
- Đăng nhập.
- Đăng xuất.
- Refresh authentication.
- Profile.
- Avatar.
- Đổi mật khẩu.

Mật khẩu phải được hash bằng thuật toán phù hợp như Argon2id hoặc bcrypt.

Nếu dùng token, ưu tiên access token ngắn hạn kết hợp refresh token có thể thu hồi.

Đối với web, dùng cookie HttpOnly, Secure trong production và SameSite phù hợp.

Không lưu access token trong localStorage nếu có thể tránh.

Bảo vệ WebSocket handshake bằng cơ chế xác thực.

Kiểm tra quyền truy cập cuộc trò chuyện ở mọi endpoint.

Không cho phép người dùng truy cập attachment của cuộc trò chuyện mà họ không phải thành viên.

## 7. POSTGRESQL DATABASE

Thiết kế database gồm các bảng chính:

### users

- id
- username
- display_name
- password_hash
- avatar_key
- created_at
- updated_at

### conversations

- id
- type: direct/group
- name
- created_by
- created_at
- updated_at
- last_message_at

### conversation_members

- conversation_id
- user_id
- role
- joined_at
- last_read_message_id
- last_read_at

### messages

- id
- conversation_id
- sender_id
- client_message_id
- type
- text_content
- reply_to_id
- created_at
- edited_at
- deleted_at

### attachments

- id
- message_id
- storage_key
- original_filename
- mime_type
- size_bytes
- width
- height
- duration_seconds
- thumbnail_key
- created_at

### media_jobs

- id
- message_id
- source_url
- provider
- status
- attempts
- error_code
- error_message
- created_at
- started_at
- completed_at

### message_reactions

- message_id
- user_id
- emoji
- created_at

### message_receipts

- message_id
- user_id
- delivered_at
- read_at

Có thể điều chỉnh schema nếu xác định được cách tối ưu hơn.

Yêu cầu:

- UUID hoặc sortable IDs phù hợp.
- Foreign keys.
- Unique constraints.
- Transaction boundaries.
- Index theo conversation_id và thứ tự tin nhắn.
- Migration có thể chạy lại an toàn theo quy trình.
- Không sử dụng offset pagination cho lịch sử chat.

Tất cả timestamp lưu dưới UTC.

Đảm bảo ID và timestamp đủ để tạo thứ tự tin nhắn ổn định khi nhiều tin được gửi cùng thời điểm.

## 8. HỆ THỐNG CHAT REAL-TIME

Sử dụng WebSocket cho các sự kiện real-time.

Các event chính:

- message.created
- message.updated
- message.deleted
- message.read
- typing.started
- typing.stopped
- conversation.updated
- attachment.ready
- media.processing
- media.completed
- media.failed

Mỗi event cần có:

- event_id
- event_type
- conversation_id
- payload
- timestamp

Có cơ chế versioning cho event payload khi cần.

### Yêu cầu vận hành

- Một user có thể mở nhiều tab.
- Một user có thể tham gia nhiều conversation.
- Một conversation có nhiều user online.
- Có heartbeat ping/pong.
- Có connection timeout.
- Có reconnect với exponential backoff.
- Có giới hạn outgoing queue.
- Không để một client chậm làm nghẽn toàn hệ thống.
- Không để nhiều goroutine ghi đồng thời vào một WebSocket connection.
- Không giữ database transaction trong khi gửi WebSocket event.

Tin nhắn phải được lưu thành công trong PostgreSQL trước khi thông báo là đã được gửi.

Khi WebSocket bị mất kết nối, frontend phải có khả năng đồng bộ lại các tin bị bỏ lỡ thông qua API.

Thiết kế cơ chế idempotency dùng client_message_id để tránh tạo tin nhắn trùng khi retry.

Với một backend instance, có thể dùng in-memory connection hub.

Nếu triển khai nhiều instance, bổ sung Redis Pub/Sub và chiến lược đảm bảo đồng bộ lại từ database.

Redis Pub/Sub không được coi là nơi lưu trữ tin nhắn bền vững.

## 9. GỬI TIN NHẮN

Hỗ trợ:

- Text.
- Emoji.
- Image.
- Video.
- File.
- Link.
- Reply.
- Edit.
- Delete.
- Reaction.

Tin nhắn được gửi qua HTTP API hoặc WebSocket theo một luồng thống nhất được thiết kế rõ ràng.

Không tạo hai con đường xử lý business logic độc lập.

Frontend sử dụng optimistic UI.

Khi người dùng nhấn gửi:

1. Tin xuất hiện ngay với trạng thái pending.
2. Client gửi request với client_message_id.
3. Backend ghi vào PostgreSQL.
4. Backend xác nhận message ID thật.
5. Frontend thay trạng thái pending thành sent.
6. WebSocket phát sự kiện tới các thành viên.

Nếu request thất bại, hiển thị retry mà không làm mất nội dung.

Không cho phép retry tạo message trùng.

## 10. UPLOAD FILE LÊN CLOUDFLARE R2

Cloudflare R2 là storage chính cho:

- Image.
- Video.
- Audio.
- File attachments.
- Thumbnail.
- Avatar.

Không lưu nội dung file trực tiếp trong PostgreSQL.

Database chỉ lưu metadata và object key.

Sử dụng AWS SDK for Go v2 với S3-compatible endpoint của R2.

Các biến môi trường:

```env
R2_ACCOUNT_ID=
R2_ACCESS_KEY_ID=
R2_SECRET_ACCESS_KEY=
R2_BUCKET_NAME=
R2_ENDPOINT=
R2_PUBLIC_BASE_URL=
```

Không hardcode credentials.

### Direct upload

Đối với file được chọn từ thiết bị người dùng:

1. Frontend gửi yêu cầu tạo upload session.
2. Backend xác thực quyền và kiểm tra quota.
3. Backend tạo storage key an toàn.
4. Backend trả presigned URL.
5. Frontend upload trực tiếp lên R2.
6. Frontend gọi endpoint hoàn tất upload.
7. Backend xác minh object thực sự tồn tại.
8. Backend kiểm tra metadata và trạng thái validation.
9. File được gắn vào message.
10. WebSocket thông báo attachment đã sẵn sàng.

Với file nhỏ, sử dụng presigned PUT.

Với file lớn, triển khai multipart upload nếu cần.

Không truyền file lớn qua Gin API nếu có thể upload trực tiếp lên R2.

Xử lý:

- Upload progress.
- Retry.
- Upload cancellation.
- Expired presigned URL.
- Upload không hoàn tất.
- Orphan objects.
- Object cleanup.

Không tin tưởng MIME type do client cung cấp.

Áp dụng giới hạn dung lượng và định dạng được phép.

### File access

Mặc định R2 bucket phải private.

Chỉ cấp quyền tải file sau khi backend xác thực quyền truy cập conversation.

Sử dụng short-lived signed GET URL hoặc phương án phân phối file có kiểm soát tương đương.

Không lưu signed URL vào database.

Không để lộ R2 secret credentials cho frontend.

## 11. HỆ THỐNG TỰ ĐỘNG TẢI VIDEO TỪ LINK

ĐÂY LÀ MỘT TRONG NHỮNG TÍNH NĂNG QUAN TRỌNG NHẤT.

Khi người dùng gửi một URL đến video trên:

- YouTube.
- Facebook.
- TikTok.
- Instagram.

Ứng dụng sẽ nhận diện link, kiểm tra tính hợp lệ và tự động xử lý bằng yt-dlp nếu nội dung được phép tải, không cần người dùng upload thủ công.

Mục tiêu cuối cùng:

**Video xuất hiện ngay trong đoạn chat dưới dạng video file có thể phát trực tiếp, thay vì chỉ là một đường link ngoài.**

Chỉ tải nội dung công khai hoặc nội dung mà người dùng có quyền tải và chia sẻ.

Không tìm cách vượt DRM, cơ chế giới hạn truy cập, quyền riêng tư hoặc xác thực của nền tảng.

Nếu video yêu cầu quyền truy cập không có sẵn, hoặc nền tảng không cho phép tải, phải trả lỗi rõ ràng và giữ link gốc.

### 11.1. URL detection

Tạo service phát hiện URL trong tin nhắn.

Cần:

- Trích xuất URL hợp lệ.
- Hỗ trợ URL có query parameters.
- Nhận diện YouTube Shorts, youtu.be, TikTok, Facebook Reels, Instagram Reels.
- Chuẩn hóa hostname và URL.
- Không nhầm URL thông thường thành video.
- Không tự tải toàn bộ playlist.
- Không tải video chỉ vì URL xuất hiện trong một đoạn văn dài nếu chưa đáp ứng quy tắc sản phẩm.

Mặc định, khi gửi một URL video độc lập thì tự động bắt đầu download.

Nếu tin nhắn có nhiều URL, hiển thị danh sách preview và tránh tạo hàng loạt download không kiểm soát.

Chỉ cho phép những platform nằm trong allowlist.

### 11.2. Luồng xử lý bất đồng bộ

Khi nhận URL:

1. Validate URL.
2. Xác thực conversation membership.
3. Tạo message type `external_video`.
4. Lưu source URL.
5. Tạo media_job với status `queued`.
6. Ghi message vào PostgreSQL.
7. Đưa job vào Redis queue.
8. Trả response ngay cho client.
9. Worker xử lý download riêng.
10. Sau khi hoàn tất, upload video lên R2.
11. Update attachment và message.
12. Phát WebSocket event.
13. Frontend đổi giao diện từ processing sang playable video.

Không chạy yt-dlp đồng bộ trong Gin handler.

Không giữ HTTP request mở trong toàn bộ quá trình tải.

Nếu enqueue thất bại, phải ghi nhận trạng thái cần retry hoặc dùng transactional outbox/reconciliation để tránh message bị kẹt vĩnh viễn.

### 11.3. yt-dlp worker

Worker Go phải gọi yt-dlp dưới dạng process riêng.

Sử dụng `exec.CommandContext` với danh sách arguments cố định và URL đã kiểm tra.

Tuyệt đối không tạo lệnh shell bằng cách nối chuỗi input người dùng.

Cần:

- Kiểm tra yt-dlp tồn tại.
- Kiểm tra FFmpeg.
- Kiểm tra FFprobe.
- Giới hạn timeout.
- Giới hạn dung lượng.
- Giới hạn video duration.
- Giới hạn số job đồng thời.
- Sử dụng thư mục tạm riêng cho mỗi job.
- Xóa file tạm khi hoàn tất hoặc thất bại.
- Capture stderr có giới hạn.
- Phân loại lỗi.
- Retry các lỗi tạm thời có backoff.

Ưu tiên giới hạn:

- Video tối đa 10 phút cho MVP.
- File tối đa 200 MB.
- Độ phân giải mục tiêu tối đa 720p.
- Tối đa 2 tác vụ download đồng thời cho môi trường local.

Các giá trị trên phải cấu hình được qua environment variables.

Không mặc định rằng metadata từ yt-dlp luôn chính xác.

Theo dõi kích thước file thực tế trong quá trình download để tránh vượt disk quota.

Không chạy download không giới hạn.

Không dùng cookies trình duyệt của người vận hành server.

### 11.4. Metadata

Trước khi tải, lấy metadata nếu có thể:

- Title.
- Duration.
- Source provider.
- Thumbnail.
- Uploader.
- Video dimensions.
- Estimated file size.

Không được tin tưởng metadata như một nguồn dữ liệu an toàn.

Đặt giới hạn độ dài chuỗi và xác thực dữ liệu trả về.

Nếu metadata không đầy đủ nhưng URL hợp lệ, có thể tiếp tục theo chính sách đã định.

### 11.5. Video processing

Sau khi tải xong:

- Dùng FFprobe để xác định codec, duration, kích thước.
- Tạo thumbnail.
- Chuẩn hóa video sang định dạng tương thích trình duyệt khi cần.
- Ưu tiên MP4 H.264 + AAC cho MVP.
- Không transcode lại nếu file đã tương thích và hợp lệ.
- Giữ tỷ lệ khung hình.
- Không upscale video không cần thiết.
- Giới hạn tài nguyên CPU, thời gian xử lý và kích thước output.

Video sau xử lý được upload lên R2.

Metadata được lưu vào PostgreSQL.

Frontend nhận thông tin để phát video bằng HTML5 `<video>`.

### 11.6. Bảo mật media download

Phải phòng chống SSRF và resource exhaustion.

Áp dụng:

- Allowlist hostname chính thức.
- Reject localhost, loopback, private IP, link-local và metadata endpoints.
- Kiểm tra DNS và redirect policy.
- Không nhận arbitrary extractor arguments từ user.
- Không nhận đường dẫn output từ user.
- Không cho yt-dlp sử dụng plugin hoặc config tùy ý từ môi trường chưa được kiểm soát.
- Chạy media worker trong container có quyền hạn thấp.
- Giới hạn egress của worker ở mức hạ tầng.
- Giới hạn disk, memory, CPU và số process.
- Không mount Docker socket.
- Không cấp R2 credentials có quyền cao hơn cần thiết.
- Dùng temp directory được quản lý.
- Không trả stderr nội bộ trực tiếp cho user.

Lưu ý rằng validate URL tại API không đủ để bảo vệ những request phụ phát sinh từ extractor.

Phải có egress controls tại tầng network/container hoặc môi trường sandbox thích hợp.

### 11.7. Error handling

Các trạng thái:

- queued
- fetching_metadata
- downloading
- processing
- uploading
- completed
- failed
- cancelled

Nếu lỗi:

- Giữ message gốc.
- Giữ source URL.
- Hiển thị trạng thái failed.
- Cung cấp retry khi phù hợp.
- Không xóa tin nhắn.
- Không retry vô hạn.
- Không tạo video attachment trùng.

Retry phải idempotent.

Nếu job chạy lại sau khi upload thành công, không upload trùng vô hạn.

Thực hiện cơ chế cleanup file tạm, multipart upload còn dang dở và object mồ côi.

### 11.8. Video trong chat

Khi video sẵn sàng, render:

- Thumbnail.
- Nút Play.
- Video duration.
- Title.
- Source platform.
- File size nếu cần.
- Trạng thái download khi chưa hoàn tất.

Click Play để phát trực tiếp trong tin nhắn.

Không tự động phát tất cả video.

Không tự động tải trước hàng trăm video.

Chỉ khởi tạo video player hoặc tải metadata khi message nằm gần viewport.

Chỉ cho phép số lượng video đang phát cùng lúc theo chính sách rõ ràng, mặc định một video tại một thời điểm.

Khi video rời viewport, cân nhắc pause và giải phóng tài nguyên phù hợp.

## 12. VIRTUALIZED CHAT — YÊU CẦU HIỆU NĂNG CỰC KỲ QUAN TRỌNG

Ứng dụng phải có khả năng hiển thị lịch sử hàng chục nghìn tin nhắn mà vẫn cuộn mượt.

Không render toàn bộ lịch sử tin nhắn trong DOM.

Sử dụng TanStack Virtual.

### 12.1. Virtual message list

Chỉ render:

- Các message đang nằm trong viewport.
- Một số message overscan ở trên và dưới.

Sử dụng ID bền vững của message làm React key.

Không sử dụng array index làm key khi danh sách có thể thêm tin cũ phía trên.

Message có chiều cao động.

Các loại tin nhắn:

- Text ngắn.
- Text nhiều dòng.
- Image.
- Video.
- File.
- Reply.
- System message.

Phải hỗ trợ đo chiều cao thực tế bằng API virtualizer phù hợp.

### 12.2. Scroll anchoring

Đây là yêu cầu bắt buộc.

Khi người dùng cuộn lên và tải thêm 30 tin cũ:

- Tin cũ được prepend vào danh sách.
- Vị trí đang đọc không thay đổi.
- Không nhảy về đầu.
- Không nhảy xuống cuối.
- Không có layout shift đáng kể.

Nếu phiên bản TanStack Virtual hỗ trợ end anchoring, áp dụng đúng API của phiên bản được cài.

Nếu không, xây dựng thuật toán bảo toàn scroll anchor dựa trên message ID, vị trí pixel và kích thước nội dung.

Không dựa vào setTimeout tùy tiện để sửa lỗi cuộn.

### 12.3. Auto scroll

Khi người dùng đang ở gần cuối:

- Tin mới xuất hiện thì tự cuộn xuống.
- Tin của chính người dùng được gửi thì có thể chủ động cuộn xuống.

Khi người dùng đang đọc lịch sử ở phía trên:

- Không tự động kéo xuống cuối.
- Hiển thị nút thông báo số tin nhắn mới.
- Click vào nút để tới tin mới nhất.

### 12.4. Dynamic media height

Ảnh và video có thể làm thay đổi chiều cao message.

Để tránh giật:

- Lưu width và height trong attachment metadata.
- Tính aspect ratio trước khi ảnh/video tải.
- Reserve trước không gian hiển thị.
- Dùng placeholder hoặc skeleton đúng tỷ lệ.
- Re-measure khi kích thước thực thay đổi.
- Bảo toàn scroll anchor.

Không để video load xong rồi đẩy toàn bộ các message khác gây nhảy màn hình.

### 12.5. Rendering optimization

Áp dụng:

- React.memo cho component phù hợp.
- Stable callbacks.
- Selector state theo từng message.
- Không rerender toàn bộ MessageList mỗi khi một người đang nhập.
- Không rebuild toàn bộ mảng message mỗi lần có typing event.
- Không decode ảnh/video ngoài viewport không cần thiết.
- Lazy-loading media.
- Abort request không còn cần thiết.
- Cleanup observer và event listener.
- Giới hạn số lượng video element đang mount.

Không lạm dụng useMemo/useCallback nếu không có lý do hiệu năng rõ ràng.

### 12.6. Pagination

Dùng cursor-based pagination.

Ví dụ:

`GET /api/v1/conversations/:id/messages?before=<cursor>&limit=30`

Có thể hỗ trợ:

- before
- after
- limit
- cursor

Backend trả:

- messages
- next_cursor
- has_more

Pagination phải có thứ tự ổn định và index phù hợp.

Frontend chỉ fetch trang cũ khi người dùng cuộn gần đầu danh sách.

Không fetch lặp cùng một page.

Dùng TanStack Query infinite query.

### 12.7. Performance targets

Thiết lập mục tiêu:

- 10.000 tin nhắn trong một conversation vẫn cuộn mượt.
- Không render 10.000 DOM message cùng lúc.
- Không bị giật nghiêm trọng khi prepend lịch sử.
- Không nhảy scroll khi ảnh/video tải xong.
- Tin nhắn mới xuất hiện ngay qua optimistic UI.
- Không bị freeze giao diện khi video đang download phía backend.

Đây là mục tiêu cần benchmark, không được tự tuyên bố đã đạt nếu chưa đo.

## 13. API DESIGN

Tạo REST API với prefix:

`/api/v1`

Nhóm endpoint dự kiến:

### Auth

- POST /auth/register
- POST /auth/login
- POST /auth/refresh
- POST /auth/logout
- GET /auth/me

### Users

- GET /users/search
- GET /users/:id
- PATCH /users/me

### Conversations

- GET /conversations
- POST /conversations
- GET /conversations/:id
- GET /conversations/:id/members
- POST /conversations/:id/members

### Messages

- GET /conversations/:id/messages
- POST /conversations/:id/messages
- PATCH /messages/:id
- DELETE /messages/:id
- POST /messages/:id/reactions
- POST /conversations/:id/read

### Attachments

- POST /attachments/uploads
- POST /attachments/uploads/:id/complete
- GET /attachments/:id/access
- POST /attachments/uploads/:id/abort

### Media

- GET /media/jobs/:id
- POST /media/jobs/:id/retry
- POST /media/jobs/:id/cancel

### WebSocket

- GET /ws

### Operations

- GET /health/live
- GET /health/ready

Các endpoint phải:

- Có authentication khi cần.
- Có authorization.
- Validate request.
- Trả error response nhất quán.
- Có pagination.
- Có request timeout.
- Có rate limiting.
- Không lộ thông tin nhạy cảm.

Viết tài liệu OpenAPI.

## 14. FRONTEND UI/UX

Thiết kế giao diện theo phong cách ứng dụng nhắn tin desktop hiện đại.

Layout:

- Sidebar danh sách cuộc trò chuyện.
- Search.
- Chat header.
- Virtualized message viewport.
- Message composer.
- Attachment preview.
- Profile/settings.

Responsive cho mobile.

### Message composer

Hỗ trợ:

- Text input nhiều dòng.
- Enter gửi.
- Shift + Enter xuống dòng.
- Emoji.
- Attach file.
- Drag and drop.
- Paste ảnh.
- Nhận diện video URL.
- Hiển thị upload progress.
- Retry gửi lỗi.

### Message bubble

Phân biệt:

- Tin của bản thân.
- Tin của người khác.
- Tin nhóm.
- Tin hệ thống.

Có:

- Timestamp.
- Sending/sent/failed.
- Read status.
- Reply.
- Reaction.
- Attachment.
- Video processing status.

### Media viewer

Hình ảnh mở lightbox.

Video phát inline.

File có nút tải xuống.

Không tự động download toàn bộ nội dung attachment chỉ vì người dùng mở conversation.

## 15. SECURITY

Bắt buộc triển khai:

- Authentication và authorization.
- CORS allowlist.
- CSRF protection nếu sử dụng cookie authentication.
- WebSocket Origin validation.
- Rate limiting.
- File size limit.
- MIME validation.
- Input validation.
- Parameterized SQL queries.
- Chống path traversal.
- Chống command injection.
- Chống SSRF.
- Chống IDOR.
- Không lộ R2 private keys.
- Không lưu password plaintext.
- Không ghi credentials hoặc token vào logs.
- Security headers.
- Giới hạn WebSocket message size.
- Giới hạn số kết nối/user.
- Giới hạn upload và media download/user.
- Kiểm tra quyền trước khi cấp URL tải attachment.

Các nội dung không đáng tin cậy như message text, video title, filename phải được escape khi render.

Không render HTML từ tin nhắn bằng dangerouslySetInnerHTML.

Không coi video download thành công đồng nghĩa với file an toàn.

Thiết kế bước xác minh/scan media phù hợp với môi trường triển khai.

## 16. DOCKER COMPOSE

Tạo Docker Compose cho local development.

Các service:

- postgres
- redis
- api
- worker
- frontend

Worker image phải cài:

- yt-dlp
- ffmpeg
- ffprobe

Sử dụng các phiên bản được pin hợp lý.

Không dùng tag `latest` tùy tiện cho môi trường production.

Cấu hình worker chạy bằng non-root user.

Thư mục tạm phải có giới hạn và được cleanup.

R2 dùng dịch vụ Cloudflare thật thông qua credentials khi có.

Cung cấp storage adapter hoặc chế độ local tương thích S3, ví dụ MinIO, để test mà không cần credentials Cloudflare thực.

Tài liệu phải nói rõ khi nào đang dùng local storage và khi nào dùng R2.

## 17. ENVIRONMENT VARIABLES

Tạo `.env.example`.

Ví dụ:

```env
APP_ENV=development
API_PORT=8080

DATABASE_URL=postgres://streamchat:password@localhost:5432/streamchat?sslmode=disable

REDIS_ADDR=localhost:6379

JWT_SECRET=
COOKIE_SECURE=false

R2_ACCOUNT_ID=
R2_ACCESS_KEY_ID=
R2_SECRET_ACCESS_KEY=
R2_BUCKET_NAME=
R2_ENDPOINT=
R2_PUBLIC_BASE_URL=

YTDLP_BINARY=yt-dlp
FFMPEG_BINARY=ffmpeg
FFPROBE_BINARY=ffprobe

MEDIA_MAX_SIZE_MB=200
MEDIA_MAX_DURATION_SECONDS=600
MEDIA_MAX_HEIGHT=720
MEDIA_DOWNLOAD_TIMEOUT_SECONDS=300
MEDIA_WORKER_CONCURRENCY=2

FRONTEND_ORIGIN=http://localhost:5173
```

Không commit `.env` thật.

Tất cả config phải được validate khi startup.

Các giới hạn download, timeout, upload và worker concurrency phải cấu hình được.

Nếu thiếu credentials R2, báo lỗi rõ ràng hoặc chạy storage adapter local khi được cấu hình.

## 18. TESTING

### Backend unit tests

- Auth.
- Conversation membership.
- Message validation.
- Pagination.
- Idempotency.
- Upload state transitions.
- Media job state transitions.
- URL validation.
- Storage permissions.

### Integration tests

- PostgreSQL migrations.
- Tạo conversation.
- Gửi và nhận tin.
- Đồng bộ WebSocket.
- Upload file.
- R2-compatible storage.
- Job queue.
- Media completion events.

### Media worker tests

Mock yt-dlp process runner.

Kiểm tra:

- URL hợp lệ và không hợp lệ.
- Timeout.
- Download thất bại.
- Metadata lỗi.
- File quá lớn.
- Video quá dài.
- Upload lỗi.
- Retry.
- Cancellation.
- Duplicate jobs.
- Temp-file cleanup.

Không phụ thuộc vào việc download video thật từ nền tảng bên ngoài để chạy toàn bộ test suite.

Có thể cung cấp smoke test tùy chọn cho video công khai có quyền sử dụng hợp lệ.

### Frontend tests

Kiểm thử:

- Virtualized list.
- Prepend history.
- Scroll anchoring.
- New-message behavior.
- Dynamic image/video height.
- Media loading/failed/completed.
- Reconnect.
- Optimistic messages.
- Upload progress.

### Performance tests

Tạo mock conversation với:

- 1.000 messages.
- 10.000 messages.
- 50.000 messages.

Đo DOM node count, render duration, scroll stability, số lần rerender và memory usage.

Không chỉ dựa vào nhận xét chủ quan rằng giao diện mượt.

## 19. THỨ TỰ TRIỂN KHAI

### Phase 1 — Project Bootstrap

- Tạo monorepo.
- Khởi tạo Go Modules.
- Khởi tạo React/Vite.
- Tạo Docker Compose.
- Cấu hình PostgreSQL, Redis.
- Health endpoints.
- Kiểm tra build.

### Phase 2 — Database & Auth

- Migrations.
- User model.
- Authentication.
- Authorization.
- Seed test users.

### Phase 3 — Messaging Core

- Conversation.
- Message persistence.
- HTTP APIs.
- WebSocket.
- Typing.
- Read status.
- Optimistic UI.

### Phase 4 — Virtualized Chat

- TanStack Virtual.
- Infinite history.
- Scroll anchoring.
- Dynamic message heights.
- React rendering optimization.

Ưu tiên hoàn thiện phase này trước khi bổ sung nhiều media component.

### Phase 5 — R2 Upload

- Storage abstraction.
- Presigned upload.
- Upload completion.
- Attachment validation.
- Image/video/file messages.

### Phase 6 — yt-dlp Pipeline

- URL detection.
- Media job persistence.
- Redis queue.
- Worker.
- Download.
- FFprobe.
- FFmpeg.
- Thumbnail.
- Upload R2.
- WebSocket status updates.

### Phase 7 — UX & Reliability

- Reconnect.
- Retry.
- Error handling.
- Security improvements.
- Tests.
- Performance benchmarks.
- Documentation.

Sau mỗi phase phải build và test.

Không chuyển sang xây dựng nhiều tính năng phụ nếu phần core chưa ổn định.

## 20. ACCEPTANCE CRITERIA

Dự án đạt MVP khi:

- Đăng ký/đăng nhập được.
- Tạo chat cá nhân và nhóm.
- Hai trình duyệt trao đổi tin real-time.
- Tin nhắn được lưu trong PostgreSQL.
- Reconnect không làm mất tin.
- Upload ảnh/video/file lên R2.
- Truy cập file có kiểm tra quyền.
- Gửi link video được hỗ trợ sẽ tạo media job.
- Worker sử dụng yt-dlp để tải video hợp lệ.
- Video được xử lý và upload lên R2.
- Tin nhắn tự chuyển sang video player khi hoàn tất.
- Job lỗi không làm mất message.
- Virtualized chat hoạt động với lịch sử dài.
- Tải tin cũ không làm nhảy scroll.
- Video load không gây layout shift đáng kể.
- Các test cốt lõi chạy thành công.
- Docker Compose có thể khởi động môi trường local.
- Có hướng dẫn chạy và triển khai.

Phân biệt rõ những gì đã thực hiện, đã build, đã test và chưa thể xác minh.

## 21. HƯỚNG DẪN THỰC THI

Bắt đầu bằng việc kiểm tra repository và môi trường.

Không xóa file hiện có nếu chưa kiểm tra.

Tạo source code thực tế.

Không chỉ lập kế hoạch rồi dừng lại.

Ưu tiên hoàn thiện từng vertical slice có thể chạy được từ frontend đến database.

Sau mỗi phase:

1. Build.
2. Chạy test.
3. Sửa lỗi.
4. Cập nhật tài liệu.
5. Chuyển sang phase tiếp theo.

Nếu thiếu credentials Cloudflare R2, dùng adapter local cho development và ghi rõ cách cấu hình R2 sau đó.

Không giả lập rằng các tích hợp bên ngoài đã hoạt động nếu chưa kiểm chứng.

Không tự ý yêu cầu tài khoản mạng xã hội, cookies đăng nhập hoặc credentials bên thứ ba.

Cuối cùng báo cáo:

- Cấu trúc dự án.
- Danh sách tính năng đã xong.
- Lệnh chạy.
- Kết quả test.
- Những lỗi còn tồn tại.
- Những tính năng chưa hoàn tất.
- Định hướng mở rộng.

**ƯU TIÊN HÀNG ĐẦU:**

1. Chat real-time ổn định và dữ liệu nhất quán.
2. Virtualized message list cuộn mượt, không giật.
3. Upload và playback media an toàn.
4. yt-dlp background pipeline không ảnh hưởng tới chat.
5. Code dễ bảo trì và có khả năng mở rộng.
