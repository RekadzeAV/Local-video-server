#include "anpr_processor.hpp"

namespace vigilos::analytics {

bool AnprProcessor::Initialize() { return true; }

bool AnprProcessor::ProcessFrame(const unsigned char* data, int width, int height) {
    (void)data;
    (void)width;
    (void)height;
    return true;
}

}  // namespace vigilos::analytics

