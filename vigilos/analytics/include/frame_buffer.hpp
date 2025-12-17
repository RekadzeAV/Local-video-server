#pragma once

namespace vigilos::analytics {
class FrameBuffer {
public:
    bool Write(const unsigned char* data, int size);
};
}  // namespace vigilos::analytics

