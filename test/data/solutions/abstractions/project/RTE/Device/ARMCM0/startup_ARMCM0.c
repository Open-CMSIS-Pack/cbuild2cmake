// expect optimize: size

#ifdef __OPTIMIZE__
#error "__OPTIMIZE__ was defined"
#endif

#ifdef __OPTIMIZE_SIZE__
#error "__OPTIMIZE_SIZE__ was defined"
#endif
