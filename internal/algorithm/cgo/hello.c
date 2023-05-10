#include <stdio.h>
#include <windows.h>
#include "hello.h"
#include <time.h>

void Hello() {
    /* 1、Linux下直接使用sleep()即可
    2、Windows下包含头文件 #include <windows.h> ，然后使用Sleep()函数，
    参数为毫秒，注意Sleep()中的S是大写
    */
    printf("start!\n");
    Sleep(3000);
//    sleep(3);
    // 利用系统时间来改变系统的种子值
//    srand(time(NULL));
//    int r = rand() % 2;
//    if (r == 0) {
//        abort();
//    }
    abort();
    printf("Hello, world!\n");
}