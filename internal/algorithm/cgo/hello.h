//#ifndef HELLO_H
//#define HELLO_H
//
//void Hello(void);
//
//#endif /* HELLO_H */

#ifndef HELLO_H
#define HELLO_H

#ifdef _WIN32
#define EXPORT __declspec(dllexport)
#else
#define EXPORT
#endif

EXPORT void Hello(void);

#endif

