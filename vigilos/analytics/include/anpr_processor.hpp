#pragma once

namespace vigilos::analytics {
class AnprProcessor {
public:
    bool Initialize();
    bool ProcessFrame(const unsigned char* data, int width, int height);
};
}  // namespace vigilos::analytics

