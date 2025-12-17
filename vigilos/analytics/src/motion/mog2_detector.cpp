#include "motion_detector.hpp"

namespace vigilos::analytics {

bool MotionDetector::Initialize() { return true; }

bool MotionDetector::ProcessFrame(const unsigned char* data, int width, int height) {
    (void)data;
    (void)width;
    (void)height;
    return true;
}

}  // namespace vigilos::analytics

