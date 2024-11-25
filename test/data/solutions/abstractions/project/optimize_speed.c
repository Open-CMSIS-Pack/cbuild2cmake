// expect optimize: speed

#ifndef __OPTIMIZE__
#error "__OPTIMIZE__ was not defined"
#endif

#ifdef __OPTIMIZE_SIZE__
#error "__OPTIMIZE_SIZE__ was defined"
#endif
