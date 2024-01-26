import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { AccommodationService } from '../../service/accommodation.service';
import { Accommodation } from '../../model/accommodation';
import { NgbDate, NgbDateStruct } from '@ng-bootstrap/ng-bootstrap';
import { AuthService } from '../../service/auth.service';
import { ReservationService } from '../../service/reservation.service';
import { ResDateRange } from '../../model/dateRange';
import { DisabledDateRange } from '../../model/disabledDateRange';
import {
  FormBuilder,
  FormGroup,
  ValidationErrors,
  Validators,
} from '@angular/forms';
<<<<<<< Updated upstream
import { ProfServiceService } from '../../service/prof.service.service';
import { NotificationService } from '../../service/notification.service';
import { Notification1 } from '../../model/notification';
=======
import { ProfServiceService } from 'src/app/service/prof.service.service';
>>>>>>> Stashed changes

@Component({
  selector: 'app-accommo-info',
  templateUrl: './accommo-info.component.html',
  styleUrls: ['./accommo-info.component.css'],
})
export class AccommoInfoComponent implements OnInit {
  grades: any[] = [];
  accommodationGrades: any[] = [];
  form: FormGroup;
  accommodationForm!: FormGroup;
  formAccommodation!: FormGroup;
  constructor(
    private fb: FormBuilder,
    private route: ActivatedRoute,
    private router: Router,
    private profService: ProfServiceService,
    private accommodationService: AccommodationService,
    private authService: AuthService,
<<<<<<< Updated upstream
    private reservationService: ReservationService,
    private notificationService: NotificationService
=======
    private reservationService: ReservationService
>>>>>>> Stashed changes
  ) {
    this.form = this.fb.group({
      grade: [
        null,
        [Validators.required, Validators.min(1), Validators.max(5)],
      ],
    });
    this.formAccommodation = this.fb.group({
      grade: [
        null,
        [Validators.required, Validators.min(1), Validators.max(5)],
      ],
    });
  }

  role: string = '';
  username: string = '';
  startDate: NgbDate | null = null;
  endDate: NgbDate | null = null;

  id: number = 0;
  reservationId: string = '';
  accommodation: Accommodation = {};
  hostId: string = '';
  dateList: ResDateRange[] = [];
  blackDateList: DisabledDateRange[] = [];
  isDataEmpty = false;

  hoveredDate: NgbDate | null = null;
  fromDate: NgbDate | null = null;
  toDate: NgbDate | null = null;
  accommodationImages: any[string] = [];

  fromDisDate: NgbDate | null = null;
  toDisDate: NgbDate | null = null;

  notification: Notification1 = {
    hostId: '',
    description: ''
  };

  ngOnInit(): void {
    this.accommodationForm = this.fb.group({
      availableFrom: ['', Validators.required],
      availableUntil: ['', Validators.required],
      pricePerNight: ['', Validators.required],
      pricePerPerson: ['', Validators.required],
    });

    this.route.params.subscribe((params) => {
      this.id = params['id'];
    });
    this.role = this.authService.getRole();
    this.username = this.authService.getUsername();

    this.accommodationService.getById(this.id).subscribe({
      next: (data) => {
        this.accommodation = data;
        console.log(this.accommodation);
        for (const image of data.images) {
          console.log(image);
          this.accommodationService.getAccommodationImage(image).subscribe(
            (blob: Blob) => {
              console.log('Blob:', blob);
              const reader = new FileReader();
              reader.onloadend = () => {
                const dataUrl = reader.result as string;
                // Now 'dataUrl' contains the data URL representation of the image
                // You can use 'dataUrl' as needed in your application
                this.accommodationImages.push(dataUrl);
                // console.log(dataUrl);
              };
              reader.readAsDataURL(blob);
            },
            (error) => {
              console.error('Error fetching image:', error);
            }
          );
        }
        // console.log(this.accommodation._id);
        // console.log(data);
      },
      error: (err) => {
        alert(err.error.message);
        console.log(err);
        this.isDataEmpty = true;
        this.router.navigate(['']);
      },
    });

    this.reservationService
      .getAvailabelDatesForAccomodation(this.id)
      .subscribe({
        next: (data) => {
          console.log(data);
          this.reservationId = data[0].reservationId;
          this.hostId = data[0].userId;
          console.log('HostId1:', this.hostId);
          this.profService.getAllHostGrades(this.hostId).subscribe({
            next: (data) => {
              console.log(data);
              this.grades = data;
            },
            error: (err) => {
              alert(err.error.message);
            },
          });
          this.accommodationService
            .getAllAccommodationGrades(this.id)
            .subscribe({
              next: (data) => {
                console.log(data);
                this.accommodationGrades = data;
              },
              error: (err) => {
                alert(err.error.message);
              },
            });
          this.startDate = new NgbDate(
            new Date(data[0].startDate).getFullYear(),
            new Date(data[0].startDate).getUTCMonth() + 1,
            new Date(data[0].startDate).getUTCDate()
          );
          this.endDate = new NgbDate(
            new Date(data[0].endDate).getFullYear(),
            new Date(data[0].endDate).getUTCMonth() + 1,
            new Date(data[0].endDate).getUTCDate()
          );
          // this.startDate = new Date(data[0].startDate);
          // this.endDate = new Date(data[0].endDate);
          // console.log(this.startDate);
          // console.log(this.endDate);
        },
        error: (err) => {
          console.log(err);
          alert(err.error.message);
          // this.router.navigate(['']);
        },
      });
    this.reservationService.getReservations(this.id).subscribe({
      next: (data) => {
        // console.log(data);
        if (data != null) {
          this.dateList = data as ResDateRange[];
          for (let dateRange of this.dateList) {
            let startDate = new NgbDate(
              new Date(dateRange.begin_accomodation_date!).getFullYear(),
              new Date(dateRange.begin_accomodation_date!).getUTCMonth() + 1,
              new Date(dateRange.begin_accomodation_date!).getUTCDate()
            );
            let endDate = new NgbDate(
              new Date(dateRange.end_accomodation_date!).getFullYear(),
              new Date(dateRange.end_accomodation_date!).getUTCMonth() + 1,
              new Date(dateRange.end_accomodation_date!).getUTCDate()
            );
            let blackDateRange: DisabledDateRange = { startDate, endDate };
            this.blackDateList.push(blackDateRange);
          }
          console.log(this.blackDateList);
        }
      },
      error: (err) => {
        console.log(err);
        alert(err.error.messa);
      },
    });
  }

  isDisabled = (date: NgbDate, current?: { month: number }) => {
    for (let dateRange of this.blackDateList) {
      if (
        (date.after(dateRange.startDate) && date.before(dateRange.endDate)) ||
        date.equals(dateRange.startDate) ||
        date.equals(dateRange.endDate)
      ) {
        return true;
      }
    }

    return date.after(this.fromDisDate) && date.before(this.toDisDate);
  };

  onDateSelection(date: NgbDate) {
    if (this.blackDateList.length > 0) {
      for (let blackDateRange of this.blackDateList) {
        if (!this.fromDate && !this.toDate) {
          this.fromDate = date;
        } else if (
          this.fromDate &&
          !this.toDate &&
          date.after(this.fromDate) &&
          ((this.fromDate.before(blackDateRange.startDate) &&
            date.before(blackDateRange.startDate)) ||
            (this.fromDate.after(blackDateRange.endDate) &&
              date.after(blackDateRange.endDate)))
        ) {
          this.toDate = date;
        } else if (
          this.fromDate &&
          this.toDate &&
          ((this.fromDate.before(blackDateRange.startDate) &&
            date.before(blackDateRange.startDate)) ||
            (this.fromDate.after(blackDateRange.endDate) &&
              date.after(blackDateRange.endDate)))
        ) {
          continue;
        } else {
          this.toDate = null;
          this.fromDate = date;
        }
      }
    } else {
      if (!this.fromDate && !this.toDate) {
        this.fromDate = date;
      } else if (this.fromDate && !this.toDate && date.after(this.fromDate)) {
        this.toDate = date;
      } else {
        this.toDate = null;
        this.fromDate = date;
      }
    }
  }

  isHovered(date: NgbDate) {
    return (
      this.fromDate &&
      !this.toDate &&
      this.hoveredDate &&
      date.after(this.fromDate) &&
      date.before(this.hoveredDate)
    );
  }

  isInside(date: NgbDate) {
    return this.toDate && date.after(this.fromDate) && date.before(this.toDate);
  }

  isRange(date: NgbDate) {
    return (
      date.equals(this.fromDate) ||
      (this.toDate && date.equals(this.toDate)) ||
      this.isInside(date) ||
      this.isHovered(date)
    );
  }

  submitGrade() {
    this.profService.gradeHost(this.hostId, this.form.value.grade).subscribe({
      next: (data) => {
        alert('Host graded');
        this.form.reset();
        this.ngOnInit();
      },
      error: (err) => {
        alert(err.error.message);
        this.form.reset();
      },
    });
<<<<<<< Updated upstream

    this.createNotification('One of your guests gave a review on you!')
=======
>>>>>>> Stashed changes
  }

  reserveDates() {
    this.reservationService
      .createReservation(
        this.reservationId,
        this.accommodation._id!,
        this.fromDate!,
        this.toDate!
      )
      .subscribe({
        next: (data) => {
          alert('Reserved');
          this.router.navigate(['']);
        },
        error: (err) => {
          console.log(err.error.message);
          alert(err.error.message);
<<<<<<< Updated upstream
          // this.ngOnInit();
=======
          this.ngOnInit();
        },
      });
  }
  deleteHostGrade(id: any) {
    this.profService.deleteHostGrades(id).subscribe({
      next: (data) => {
        alert('Grade deleted');
        this.ngOnInit();
      },
      error: (err) => {
        alert(err.error.message);
      },
    });
  }
  deleteAccommodationGrade(id: any) {
    this.accommodationService.deleteAccommodationGrade(id).subscribe({
      next: (data) => {
        alert('Accommodation deleted');
        this.ngOnInit();
      },
      error: (err) => {
        alert(err.error.message);
      },
    });
  }

  submitAccommodationGrade() {
    this.accommodationService
      .gradeAccommodation(this.id, this.formAccommodation.value.grade)
      .subscribe({
        next: (data) => {
          alert('Accommodation graded');
          this.ngOnInit();
        },
        error: (err) => {
          alert(err.error.message);
>>>>>>> Stashed changes
        },
      });

      this.createNotification('One of your accommodations just got reserved!')
  }
  deleteHostGrade(id: any) {
    this.profService.deleteHostGrades(id).subscribe({
      next: (data) => {
        alert('Grade deleted');
        this.ngOnInit();
      },
      error: (err) => {
        alert(err.error.message);
      },
    });

    this.createNotification('One of your guests deleted their review on you!')
  }
  deleteAccommodationGrade(id: any) {
    this.accommodationService.deleteAccommodationGrade(id).subscribe({
      next: (data) => {
        alert('Accommodation deleted');
        this.ngOnInit();
      },
      error: (err) => {
        alert(err.error.message);
      },
    });
    this.createNotification('One of your guests deleted their review on one of your accommodations!')
  }

  submitAccommodationGrade() {
    this.accommodationService
      .gradeAccommodation(this.id, this.formAccommodation.value.grade)
      .subscribe({
        next: (data) => {
          alert('Accommodation graded');
          this.ngOnInit();
        },
        error: (err) => {
          alert(err.error.message);
        },
      });

     this.createNotification('One of your guests left a review on one of your accommodations!')
  }

  createNotification(description: string){
    this.notification.hostId = this.hostId;
    this.notification.description = description
    this.notificationService.createNotification(this.notification).subscribe({
      next: (data) => {
        alert('Notification Sent')
      },
      error: (err) => {
        alert(err.error.message)
      }
    })
  }

  onSubmit() {
    console.log(this.accommodation._id, this.accommodationForm.value);
    this.reservationService
      .createReservationDatesForAccomodation(
        this.accommodation._id,
        this.accommodationForm.value
      )
      .subscribe({
        next: (data) => {
          console.log('Reservation in succesfuly created');
          alert('Reservation is successfully.');
          this.router.navigate(['accommodations/myAccommodations']);
        },
        error: (err) => {
          console.log(err.error.message);
          alert(err.error.message);
        },
      });
  }
}
